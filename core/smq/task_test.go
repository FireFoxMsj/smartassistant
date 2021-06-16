package smq

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
)

var manualScene = orm.Scene{
	Name:      "manual",
	AutoRun:   false,
	CreatorID: 111,
	CreatedAt: time.Now(),
	SceneTasks: []orm.SceneTask{
		{
			ControlSceneID: 0,
			DelaySeconds:   10,
			Type:           orm.TaskTypeSmartDevice,
			SceneTaskDevices: []orm.SceneTaskDevice{
				{
					DeviceID:  22,
					Action:    "switch",
					ActionVal: "on",
				},
			},
		},
	},
}

var autoScene = orm.Scene{
	Name:           "auto",
	AutoRun:        true,
	ConditionLogic: orm.MatchAllCondition,
	RepeatType:     orm.RepeatTypeAllDay,
	RepeatDate:     "1234567",
	SceneConditions: []orm.SceneCondition{
		{
			ConditionType: orm.ConditionTypeTiming,
			TimingAt:      time.Now().Add(15 * time.Second),
			DeviceID:      1,
			ConditionItem: orm.ConditionItem{},
		},
	},
	IsOn:           true,
	TimePeriodType: orm.TimePeriodTypeAllDay,
	CreatorID:      111,
	CreatedAt:      time.Now(),
	SceneTasks: []orm.SceneTask{
		{
			DelaySeconds: 10,
			Type:         orm.TaskTypeSmartDevice,
			SceneTaskDevices: []orm.SceneTaskDevice{
				{
					DeviceID:  22,
					Action:    "switch",
					ActionVal: "off",
				},
			},
		},
		{
			ControlSceneID: 1,
			DelaySeconds:   15,
			Type:           orm.TaskTypeManualRun,
		},
	},
}

func TestAddScene(t *testing.T) {

	go MinHeapQueue.Start()
	time.Sleep(2 * time.Second)

	// 创建两个场景
	db := orm.GetDB().Session(&gorm.Session{FullSaveAssociations: true}).Model(orm.Scene{})
	db.Create(&manualScene) // 10s后开
	db.Create(&autoScene)   // 定时执行 10s后关+15s后执行场景1

	AddSceneTaskByID(2)

	time.Sleep(100 * time.Second)
}
