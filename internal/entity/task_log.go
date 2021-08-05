package entity

import (
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"time"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// TaskResultType 任务执行结果
type TaskResultType int

const (
	TaskSuccess TaskResultType = iota + 1
	TaskPartSuccess
	TaskFail
	TaskTimeout
	TaskDeviceAlreadyDeleted
	TaskDeviceDisConnect
	TaskSceneAlreadyDeleted
)

var (
	taskErrMap = map[errors.Code]TaskResultType{
		errors.GetCode(status.SceneNotExist):  TaskSceneAlreadyDeleted,
		errors.GetCode(status.DeviceNotExist): TaskDeviceAlreadyDeleted,
		errors.GetCode(status.DeviceOffline):  TaskDeviceDisConnect,
	}
)

// TaskLog 任务日志
type TaskLog struct {
	ID   int
	Name string   // 设备名称/场景名称
	Type TaskType // 任务类型

	Finish bool           // 是否完成
	Result TaskResultType // 执行结果
	Error  string

	DeviceLocation string // 设备区域

	TaskID        string    `gorm:"uniqueIndex"` // 任务ID
	ParentTaskID  *string   // 父任务id
	ChildTaskLogs []TaskLog `gorm:"foreignkey:parent_task_id;references:task_id"` // 子任务日志

	FinishedAt time.Time
	CreatedAt  time.Time
}

func (tl TaskLog) TableName() string {
	return "task_logs"
}

// UpdateTaskLog 更新任务日志
func UpdateTaskLog(taskID string, taskErr error) error {

	update := TaskLog{
		FinishedAt: time.Now(),
		Result:     TaskSuccess,
	}
	// 根据错误更新不同执行结果
	if taskErr != nil {
		update.Result = TaskFail
		if v, ok := taskErr.(errors.Error); ok { // 判断错误类型
			update.Result, _ = taskErrMap[v.Code]
		}
		update.Error = taskErr.Error()
	}
	// 更新日志
	var taskLog TaskLog
	if err := GetDB().Where("task_id=?", taskID).First(&taskLog).Updates(&update).Error; err != nil {
		return err
	}

	if taskLog.ParentTaskID != nil {
		return UpdateParentLog(*taskLog.ParentTaskID)
	}
	return nil
}

// UpdateParentLog 更新父任务的日志
func UpdateParentLog(parentTaskID string) error {

	// 父任务的所有子任务都完成则更新父任务为已完成
	var taskLogs []TaskLog
	if err := GetDB().Where("parent_task_id=?", parentTaskID).
		Find(&taskLogs).Error; err != nil {
		return err
	}

	update := TaskLog{
		Finish:     true,
		FinishedAt: time.Now(),
	}
	var errCount int
	for _, tl := range taskLogs {
		if tl.Result == 0 {
			return nil
		}
		if tl.Error != "" {
			errCount += 1
		}
	}
	if errCount == len(taskLogs) {
		update.Result = TaskFail
	} else if errCount == 0 {
		update.Result = TaskSuccess
	} else {
		update.Result = TaskPartSuccess
	}
	var parentTaskLog TaskLog
	if err := GetDB().Where("task_id=?", parentTaskID).First(&parentTaskLog).
		Updates(&update).Error; err != nil {
		return err
	}

	if parentTaskLog.ParentTaskID != nil {
		return UpdateParentLog(*parentTaskLog.ParentTaskID)
	}
	return nil
}

func NewTaskLog(target interface{}, taskID string, parentTaskID *string) error {

	var (
		name     string
		taskType TaskType
		location Location
	)
	switch v := target.(type) {
	case Scene:
		name = v.Name
		taskType = TaskTypeManualRun
		if v.AutoRun {
			taskType = TaskTypeEnableAutoRun // TODO 不准确
		}
	case Device:
		name = v.Name
		location, _ = GetLocationByID(v.LocationID)
		taskType = TaskTypeSmartDevice
	}
	taskLog := TaskLog{
		Name:           name,
		DeviceLocation: location.Name,
		Type:           taskType,
		TaskID:         taskID,
		ParentTaskID:   parentTaskID,
		CreatedAt:      time.Now(),
	}
	return GetDB().Create(&taskLog).Error
}
