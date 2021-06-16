package smq

import (
	errors2 "errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/now"
	"gorm.io/gorm"

	"gitlab.yctc.tech/root/smartassistent.git/core"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

// taskLogWrapper 包装任务执行函数，在任务执行时插入日志，任务执行完后更新日志
func taskLogWrapper(target interface{}) WrapperFunc {
	return func(f TaskFunc) TaskFunc {
		return func(task *Task) error {
			if task == nil {
				return nil
			}
			var parentID *string
			if task.Parent != nil {
				parentID = &task.Parent.ID
			}
			if s, ok := target.(orm.Scene); ok { // 记录场景正在执行的task的index
				MinHeapQueue.SceneTaskIndexMap.Store(s.ID, task.index)
			}
			if err := orm.NewTaskLog(target, task.ID, parentID); err != nil {
				log.Println("NewTaskLogErr:", err)
			}
			err := f(task)
			if e := orm.UpdateTaskLog(task.ID, err); e != nil {
				log.Println(e)
			}
			return err
		}
	}
}

func PushTask(task *Task, target interface{}) {
	task.WithWrapper(taskLogWrapper(target))
	MinHeapQueue.Push(task)
}

// DelSceneTask 删除场景正在执行的任务
func DelSceneTask(sceneID int) {
	if taskIndex, ok := MinHeapQueue.SceneTaskIndexMap.LoadAndDelete(sceneID); ok {
		MinHeapQueue.Remove(taskIndex.(int))
	}
}

// RestartSceneTask 重启场景对应的任务（就是删除然后重新添加任务）
func RestartSceneTask(sceneID int) error {
	DelSceneTask(sceneID)
	return AddSceneTaskByID(sceneID)
}

// AddSceneTaskByID 根据场景id执行场景（执行或者开启时调用）
func AddSceneTaskByID(sceneID int) error {

	scene, err := orm.GetSceneInfoById(sceneID)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.SceneNotExist)
		}
		return errors.Wrap(err, errors.InternalServerErr)
	}
	AddSceneTask(scene)
	return nil
}

// AddSceneTask 添加场景任务（执行或者开启时调用）
func AddSceneTask(scene orm.Scene) {

	var t *Task
	if scene.AutoRun { // 开启自动场景
		fmt.Printf("open scene %d\n", scene.ID)
		// 找到定时条件的时间
		for _, c := range scene.SceneConditions {
			if c.ConditionType == orm.ConditionTypeTiming {

				// 获取任务今天的下次执行时间
				days := time.Now().Sub(c.TimingAt).Hours() / 24
				nextTime := c.TimingAt.AddDate(0, 0, int(days))
				if nextTime.Before(time.Now()) || nextTime.After(now.EndOfDay()) {
					fmt.Printf("now:%v,invalid next execute time:%v", time.Now(), nextTime)
					continue
				}

				t = NewTaskAt(WrapSceneFunc(scene, true), nextTime)
				PushTask(t, scene)
				continue
			}
		}
	} else { // 执行手动场景
		fmt.Printf("execute scene %d\n", scene.ID)
		t = NewTask(WrapSceneFunc(scene, false), 0)
		PushTask(t, scene)
	}
}

// WrapSceneFunc  包装场景为任务
func WrapSceneFunc(scene orm.Scene, isTrigByTime bool) (f TaskFunc) {
	return func(t *Task) error {
		if scene.Deleted.Valid { // 已删除的场景不执行
			return errors.New(errors.SceneNotExist)
		}
		if scene.AutoRun && !IsConditionsSatisfied(scene, isTrigByTime) { // 自动场景则判断条件
			fmt.Printf("auto scene:%d's conditons not astisfied\n", scene.ID)
			return nil
		}
		for _, sceneTask := range scene.SceneTasks {
			delay := time.Duration(sceneTask.DelaySeconds) * time.Second
			task := NewTask(WrapTaskToFunc(sceneTask), delay).WithParent(t)

			if sceneTask.Type == orm.TaskTypeSmartDevice { // 控制设备
				if len(sceneTask.SceneTaskDevices) == 0 {
					continue
				}
				deviceID := sceneTask.SceneTaskDevices[0].DeviceID
				var device orm.Device
				orm.GetDB().Unscoped().First(&device, deviceID)
				PushTask(task, device)
			} else {
				controlScene, _ := orm.GetDeletedSceneByID(sceneTask.ControlSceneID)
				PushTask(task, controlScene)
			}
		}
		return nil
	}
}

// WrapTaskToFunc 包装任务为执行函数
func WrapTaskToFunc(task orm.SceneTask) (f TaskFunc) {

	return func(t *Task) error {
		// TODO 判断权限、判断场景是否有修改
		fmt.Printf("execute task:%d,type:%d\n", task.ID, task.Type)
		switch task.Type {
		case orm.TaskTypeSmartDevice: // 控制设备
			return ExecuteDevice(task.SceneTaskDevices)
		case orm.TaskTypeManualRun: // 执行场景
			return AddSceneTaskByID(task.ControlSceneID)
		case orm.TaskTypeEnableAutoRun: // 开启场景
			return SetSceneOn(task.ControlSceneID)
		case orm.TaskTypeDisableAutoRun: // 关闭场景
			return SetSceneOff(task.ControlSceneID)
		}
		return nil
	}
}

// IsConditionsSatisfied 场景条件是否满足 isTrigByTime 是否由定时条件触发
func IsConditionsSatisfied(scene orm.Scene, isTrigByTime bool) bool {
	if !scene.IsOn {
		log.Printf("scene %d: is off\n", scene.ID)
		return false
	}
	if !IsInTimePeriod(scene) { // 不在有效时间段内则不执行
		log.Printf("scene %d: not in effective time period\n", scene.ID)
		return false
	}
	// “任一满足”情况下，定时触发的任务直接满足条件
	if scene.ConditionLogic == orm.MatchAnyCondition && isTrigByTime {
		return true
	}
	for _, condition := range scene.SceneConditions {
		if condition.ConditionType == orm.ConditionTypeTiming {
			continue
		}

		// 任一满足
		if scene.ConditionLogic == orm.MatchAnyCondition && IsConditionSatisfied(condition) {
			log.Printf("scene %d: condition:%d satisfied\n", scene.ID, condition.ID)
			return true
		}
		// 全部满足（有一个不满足）
		if scene.ConditionLogic == orm.MatchAllCondition && !IsConditionSatisfied(condition) {
			log.Printf("scene %d: condition:%d not satisfied\n", scene.ID, condition.ID)
			return false
		}
	}

	if scene.ConditionLogic == orm.MatchAllCondition {
		log.Printf("scene %d: all conditions satisfied\n", scene.ID)
		return true
	} else {
		log.Printf("scene %d: no any conditions satisfied\n", scene.ID)
		return false
	}
}

// IsInTimePeriod 是否在时间段内
func IsInTimePeriod(scene orm.Scene) bool {

	weekday := time.Now().Weekday()
	if !strings.Contains(scene.RepeatDate, strconv.Itoa(int(weekday))) {
		log.Printf("scene %d: today not in repeat date\n", scene.ID)
		return false
	}

	if scene.TimePeriodType == orm.TimePeriodTypeCustom {
		days := int(time.Now().Sub(scene.EffectStart).Hours() / 24)
		effectEndTime := scene.EffectEnd.AddDate(0, 0, days)
		effectStartTime := scene.EffectStart.AddDate(0, 0, days)
		return time.Now().Before(effectEndTime) && time.Now().After(effectStartTime)
	}
	return true
}

// IsConditionSatisfied 判断设备状态是否满足条件
func IsConditionSatisfied(condition orm.SceneCondition) bool {
	if condition.ConditionType == orm.ConditionTypeTiming {
		return false
	}

	device, err := orm.GetDeviceByID(condition.DeviceID)
	if err != nil {
		log.Printf("get device %d err:%v\n", condition.DeviceID, err)
		return false
	}
	// 获取设备当前状态
	var states map[string]interface{}
	states, err = core.GetDeviceStateFromResp(device)
	if err != nil {
		return false
	}
	item := condition.ConditionItem
	if val, ok := states[item.Attribute]; ok {
		switch item.Operator {
		case orm.OperatorEQ:
			return val == item.ActionVal
		case orm.OperatorGT:
			itemVal, _ := strconv.Atoi(item.ActionVal)
			return int(val.(float64)) > itemVal
		case orm.OperatorLT:
			itemVal, _ := strconv.Atoi(item.ActionVal)
			return int(val.(float64)) < itemVal
		}
	}
	return false
}

// ExecuteDevice 控制设备执行
func ExecuteDevice(ds []orm.SceneTaskDevice) (err error) {

	for _, d := range ds {
		log.Printf("control device %d: %s %s\n", d.DeviceID, d.Action, d.ActionVal)
		device, err := orm.GetDeviceByID(d.DeviceID)
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.DeviceNotExist)
		}

		state, _ := core.GetDeviceStateFromResp(device)
		if v, ok := state["is_online"]; ok {
			if v == false {
				return errors.New(errors.DeviceOffline)
			}
		}
		action := orm.GetDeviceActionByAttr(device, d.Attribute)
		data := core.M{
			"service_name":   action.Action,
			"id":             device.Identity,
			action.Attribute: d.ActionVal}
		if err = core.Sass.Services.Call(device.Manufacturer, action.Action, data); err != nil {
			log.Println("execute device cmd err", err)
		}
	}
	return
}

// SetSceneOn 开启场景
func SetSceneOn(sceneID int) (err error) {
	if err = orm.SwitchAutoSceneByID(sceneID, true); err != nil {
		return
	}
	if err := AddSceneTaskByID(sceneID); err != nil {
		log.Println(err)
	}
	return
}

// SetSceneOff 关闭场景
func SetSceneOff(sceneID int) (err error) {
	if err = orm.SwitchAutoSceneByID(sceneID, false); err != nil {
		return
	}
	DelSceneTask(sceneID)
	return
}

// DeviceStateChange 设备状态变化触发场景
func DeviceStateChange(deviceID int, state map[string]interface{}) {
	for attr := range state { // 触发设备状态变更
		DeviceAttrChange(deviceID, attr)
	}
}

// DeviceAttrChange 设备属性变更时触发场景
func DeviceAttrChange(deviceID int, attr string) {

	scenes, err := orm.GetScenesByCondition(deviceID, attr)
	if err != nil {
		log.Printf("can't find scenes with device %d %s change", deviceID, attr)
		return
	}

	// 遍历并包装场景为任务
	for _, scene := range scenes {
		scene, _ = orm.GetSceneInfoById(scene.ID)
		// 全部满足且有定时条件则不执行
		if scene.ConditionLogic == orm.MatchAllCondition && IsSceneHaveTimeCondition(scene) {
			fmt.Printf("device %d state %s changed but scenes %d not match time conditoin,ignore\n",
				deviceID, attr, scene.ID)
			continue
		}
		t := NewTask(WrapSceneFunc(scene, false), 0)
		PushTask(t, scene)
	}
}

// IsSceneHaveTimeCondition 场景是否有定时条件
func IsSceneHaveTimeCondition(scene orm.Scene) bool {

	for _, c := range scene.SceneConditions {
		if c.ConditionType == orm.ConditionTypeTiming {
			return true
		}
	}
	return false
}

// addSceneTaskByTime 编排场景任务
func addSceneTaskByTime(t time.Time) {
	scenes, err := orm.GetPendingScenesByTime(t)
	if err != nil {
		log.Printf("get execute scenes err %v\n", err)
		return
	}

	for _, scene := range scenes {
		// 没有定时触发条件 不加入队列
		if !IsSceneHaveTimeCondition(scene) {
			continue
		}
		AddSceneTask(scene)
	}
}

// AddArrangeSceneTask 每天定时编排场景任务
func AddArrangeSceneTask(executeTime time.Time) {
	var f TaskFunc
	f = func(task *Task) error {
		addSceneTaskByTime(executeTime.AddDate(0, 0, 1))

		// 将下一个定时编排任务排进队列
		AddArrangeSceneTask(executeTime.AddDate(0, 0, 1))
		return nil
	}

	task := NewTaskAt(f, executeTime)
	MinHeapQueue.Push(task)
}

func init() {
	InitHeap()
	// 重启时编排任务
	addSceneTaskByTime(time.Now())
	// 每天 23:55:00 进行第二天任务编排
	AddArrangeSceneTask(now.EndOfDay().Add(-5 * time.Minute))
}
