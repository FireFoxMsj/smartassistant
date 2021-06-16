package orm

import (
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

// 一个任务仅允许关联一个设备，对应的多个功能点配置；
// 或者
// 一个任务仅允许控制同一场景类型下的多个场景

type TaskType int

const (
	TaskTypeSmartDevice TaskType = iota + 1
	TaskTypeManualRun
	TaskTypeEnableAutoRun
	TaskTypeDisableAutoRun
)

// SceneTask 场景任务
type SceneTask struct {
	ID               int               `json:"id"`
	SceneID          int               `json:"scene_id"`
	ControlSceneID   int               `json:"control_scene_id"` // ControlSceneID 控制场景id
	DelaySeconds     int               `json:"delay_seconds"`    // 延迟的秒数
	Type             TaskType          `json:"type"`             // 任务目标：智能设备device或者是场景scene
	SceneTaskDevices []SceneTaskDevice `json:"scene_task_devices" gorm:"constraint:OnDelete:CASCADE;"`
}

func (d SceneTask) TableName() string {
	return "scene_tasks"
}

func GetSceneTasksBySceneID(sceneID int) (sceneTasks []SceneTask, err error) {
	err = GetDB().Order("type asc").Where("scene_id = ?", sceneID).Find(&sceneTasks).Error
	return
}

// SceneTaskDevice 智能设备执行的具体功能点
type SceneTaskDevice struct {
	ID          int    `json:"id"`
	SceneTaskID int    `json:"scene_task_id"`
	DeviceID    int    `json:"device_id"`
	Action      string `json:"action"`     // 功能点，比如开关、亮度、色温
	Attribute   string `json:"attribute"`  // 功能对应的属性，比如开关电源，开关左键
	ActionVal   string `json:"action_val"` // 对应的值
}

func (d SceneTaskDevice) TableName() string {
	return "scene_task_devices"
}

func GetSceneTaskItemsByTaskID(taskID int) (taskItems []SceneTaskDevice, err error) {
	err = GetDB().Where("scene_task_id = ?", taskID).Find(&taskItems).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}

	return
}

func CreateSceneTask(sceneTask []SceneTask) (err error) {
	err = GetDB().Create(&sceneTask).Error
	if err != nil {
		err = errors.New(errors.InternalServerErr)
	}
	return
}

func CreateTaskPerformItem(performItem []SceneTaskDevice) (err error) {
	if err = GetDB().Create(&performItem).Error; err != nil {
		err = errors.New(errors.InternalServerErr)
	}
	return
}

// checkPerformItem 执行任务为控制设备时，对应操作校验
func (td SceneTaskDevice) checkPerformItem() (err error) {
	if td.DeviceID == 0 {
		err = errors.New(errors.BadRequest)
		return
	}

	if err = checkOperation(td.Action); err != nil {
		return
	}
	return

}

// checkTaskDevice 校验设备任务类型
func (task SceneTask) CheckTaskDevice(userId int) (err error) {
	if task.SceneTaskDevices == nil {
		err = errors.New(errors.DeviceOperationNotSetErr)
		return
	}

	for _, taskDevice := range task.SceneTaskDevices {
		if !IsDeviceControlPermit(userId, taskDevice.DeviceID, taskDevice.Attribute) {
			err = errors.New(errors.DeviceOrSceneControlDeny)
			return
		}
		if err = taskDevice.checkPerformItem(); err != nil {
			return
		}
	}
	return
}

// checkTaskType 执行任务类型校验
func (task SceneTask) CheckTaskType() (err error) {
	if task.Type < TaskTypeSmartDevice || task.Type > TaskTypeDisableAutoRun {
		err = errors.New(errors.TaskTypeErr)
	}
	return
}
