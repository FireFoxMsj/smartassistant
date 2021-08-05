package task

import (
	"context"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jinzhu/now"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/plugin"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	plugin2 "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	"gorm.io/gorm"
)

var (
	manager     *Manager
	managerOnce sync.Once
)

// Manager Task 服务
type Manager struct {
	queue         *QueueServe
	pluginManager *plugin.Manager
	runningScene  sync.Map // 正在执行的场景的id -> queue index
}

func GetManager() *Manager {
	managerOnce.Do(func() {
		manager = &Manager{
			queue:         NewQueueServe(),
			pluginManager: plugin.GetManager(),
		}
	})
	return manager
}

// Run 启动服务，扫描插件并且连接通讯
func (m *Manager) Run(ctx context.Context) {
	logrus.Info("starting task manager")
	go m.queue.Start(ctx)
	// 重启时编排任务
	m.addSceneTaskByTime(time.Now())
	// 每天 23:55:00 进行第二天任务编排
	m.addArrangeSceneTask(now.EndOfDay().Add(-5 * time.Minute))
	// TODO 扫描已安装的插件，并且启动，连接 state change...
	<-ctx.Done()
	// TODO 断开连接
	logrus.Warning("task manager stopped")
}

// addSceneTaskByTime 编排场景任务
func (m *Manager) addSceneTaskByTime(t time.Time) {
	scenes, err := entity.GetPendingScenesByTime(t)
	if err != nil {
		logrus.Errorf("get execute scenes err %v", err)
		return
	}
	for _, scene := range scenes {
		// 没有定时触发条件 不加入队列
		if !IsSceneHaveTimeCondition(scene) {
			continue
		}
		m.AddSceneTask(scene)
	}
}

// addArrangeSceneTask 每天定时编排场景任务
func (m *Manager) addArrangeSceneTask(executeTime time.Time) {
	var f TaskFunc
	f = func(task *Task) error {
		m.addSceneTaskByTime(executeTime.AddDate(0, 0, 1))

		// 将下一个定时编排任务排进队列
		m.addArrangeSceneTask(executeTime.AddDate(0, 0, 1))
		return nil
	}

	task := NewTaskAt(f, executeTime)
	m.PushTask(task, "daily arrange scene task")
}

// DeleteSceneTask 删除场景任务
func (m *Manager) DeleteSceneTask(sceneID int) {
	// 现时需求如果场景对应的任务已运行，则不需要处理
}

// AddSceneTaskByID 根据场景id执行场景（执行或者开启时调用）
func (m *Manager) AddSceneTaskByID(sceneID int) error {
	scene, err := entity.GetSceneInfoById(sceneID)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(status.SceneNotExist)
		}
		return errors.Wrap(err, errors.InternalServerErr)
	}
	m.AddSceneTask(scene)
	return nil
}

// AddSceneTask 添加场景任务（执行或者开启时调用）
func (m *Manager) AddSceneTask(scene entity.Scene) {
	var t *Task
	if scene.AutoRun { // 开启自动场景
		logrus.Infof("open scene %d", scene.ID)
		// 找到定时条件的时间
		for _, c := range scene.SceneConditions {
			if c.ConditionType == entity.ConditionTypeTiming {

				// 获取任务今天的下次执行时间
				execTime := now.BeginningOfDay().Add(c.TimingAt.Sub(now.New(c.TimingAt).BeginningOfDay()))
				if execTime.Before(time.Now()) || execTime.After(now.EndOfDay()) {
					logrus.Infof("now:%v,invalid next execute time:%v", time.Now(), execTime)
					continue
				}

				t = NewTaskAt(m.WrapSceneFunc(scene, true), execTime)
				m.PushTask(t, scene)
				continue
			}
		}
	} else { // 执行手动场景
		logrus.Infof("execute scene %d", scene.ID)
		t = NewTask(m.WrapSceneFunc(scene, false), 0)
		m.PushTask(t, scene)
	}
}

func (m *Manager) PushTask(task *Task, target interface{}) {
	task.WithWrapper(taskLogWrapper(target))
	m.queue.Push(task)
}

// RestartSceneTask 重启场景对应的任务（就是删除然后重新添加任务）
func (m *Manager) RestartSceneTask(sceneID int) error {
	m.DeleteSceneTask(sceneID)
	return m.AddSceneTaskByID(sceneID)
}

func (m *Manager) addRunningScene(sceneID int, queueIndex int) {
	m.runningScene.Store(sceneID, queueIndex)
}

// WrapSceneFunc  包装场景为 TaskFunc
func (m *Manager) WrapSceneFunc(sc entity.Scene, isTrigByTimer bool) (f TaskFunc) {
	return func(t *Task) error {
		scene, err := entity.GetSceneInfoById(sc.ID)
		if err != nil {
			if errors2.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(status.SceneNotExist)
			}
			return errors.Wrap(err, errors.InternalServerErr)
		}
		if scene.Deleted.Valid { // 已删除的场景不执行
			return errors.New(status.SceneNotExist)
		}
		if scene.AutoRun && !IsConditionsSatisfied(scene, isTrigByTimer) { // 自动场景则判断条件
			logrus.Infof("auto scene:%d's conditons not satisfied", scene.ID)
			return nil
		}
		// TODO 此代码达到其功能，需清理
		m.addRunningScene(scene.ID, t.index)
		for _, sceneTask := range scene.SceneTasks {
			delay := time.Duration(sceneTask.DelaySeconds) * time.Second
			task := NewTask(m.WrapTaskToFunc(sceneTask), delay).WithParent(t)

			if sceneTask.Type == entity.TaskTypeSmartDevice { // 控制设备
				if len(sceneTask.Attributes) == 0 {
					continue
				}
				deviceID := sceneTask.DeviceID
				var device entity.Device
				entity.GetDB().Unscoped().First(&device, deviceID)
				device, err := entity.GetDeviceByIDWithUnscoped(deviceID)
				if err == nil {
					m.PushTask(task, device)
				}
			} else {
				controlScene, err := entity.GetSceneByIDWithUnscoped(sceneTask.ControlSceneID)
				if err == nil {
					m.PushTask(task, controlScene)
				}
			}
		}
		return nil
	}
}

// WrapTaskToFunc 包装场景任务为 TaskFunc
func (m *Manager) WrapTaskToFunc(task entity.SceneTask) (f TaskFunc) {
	return func(t *Task) error {
		// TODO 判断权限、判断场景是否有修改
		fmt.Printf("execute task:%d,type:%d\n", task.ID, task.Type)
		switch task.Type {
		case entity.TaskTypeSmartDevice: // 控制设备
			return m.executeDevice(task)
		case entity.TaskTypeManualRun: // 执行场景
			return m.AddSceneTaskByID(task.ControlSceneID)
		case entity.TaskTypeEnableAutoRun: // 开启场景
			return m.setSceneOn(task.ControlSceneID)
		case entity.TaskTypeDisableAutoRun: // 关闭场景
			return m.setSceneOff(task.ControlSceneID)
		}
		return nil
	}
}

// executeDevice 控制设备执行
func (m *Manager) executeDevice(task entity.SceneTask) (err error) {

	var ds []entity.Attribute
	if err := json.Unmarshal(task.Attributes, &ds); err != nil {
		logrus.Error(err)
		return err
	}
	for _, d := range ds {
		var device entity.Device
		device, err = entity.GetDeviceByID(task.DeviceID)
		if err != nil {
			if errors2.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(status.DeviceNotExist)
			}
			return errors.New(http.StatusInternalServerError)
		}
		logrus.Infof("execute device command device id:%d instance id:%d attr:%s val:%v",
			device.ID, d.InstanceID, d.Attribute.Attribute, d.Attribute.Val)

		attributes := []plugin2.SetAttribute{
			{
				InstanceID: d.InstanceID,
				Attribute:  d.Attribute.Attribute,
				Val:        d.Attribute.Val,
			},
		}

		data, _ := json.Marshal(plugin2.SetRequest{Attributes: attributes})
		err = plugin.SetAttributes(device.PluginID, device.Identity, data)
		if err != nil {
			return
		}
	}
	return
}

// SetSceneOn 开启场景
func (m *Manager) setSceneOn(sceneID int) (err error) {
	if err = entity.SwitchAutoSceneByID(sceneID, true); err != nil {
		return
	}
	if err := m.AddSceneTaskByID(sceneID); err != nil {
		logrus.Error(err)
	}
	return
}

// SetSceneOff 关闭场景
func (m *Manager) setSceneOff(sceneID int) (err error) {
	if err = entity.SwitchAutoSceneByID(sceneID, false); err != nil {
		return
	}
	m.DeleteSceneTask(sceneID)
	return
}

// DeviceStateChange 设备状态变化触发场景
func (m *Manager) DeviceStateChange(identity string, attr entity.Attribute) {
	d, _ := entity.GetDeviceByIdentity(identity)
	m.DeviceAttrChange(d.ID, attr)
}

// DeviceAttrChange 设备属性变更时触发场景
func (m *Manager) DeviceAttrChange(deviceID int, attr entity.Attribute) {

	scenes, err := entity.GetScenesByCondition(deviceID, attr)
	if err != nil {
		logrus.Errorf("can't find scenes with device %d %d %s change",
			deviceID, attr.InstanceID, attr.Attribute.Attribute)
		return
	}

	// 遍历并包装场景为任务
	for _, scene := range scenes {
		scene, _ = entity.GetSceneInfoById(scene.ID)
		// 全部满足且有定时条件则不执行
		if scene.IsMatchAllCondition() && IsSceneHaveTimeCondition(scene) {
			fmt.Printf("device %d state %s changed but scenes %d not match time conditoin,ignore\n",
				deviceID, attr.Attribute.Attribute, scene.ID)
			continue
		}
		t := NewTask(m.WrapSceneFunc(scene, false), 0)
		m.PushTask(t, scene)
	}
}
