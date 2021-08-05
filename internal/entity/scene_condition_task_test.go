package entity

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/zhiting-tech/smartassistant/internal/types"
	"strconv"
	"testing"
	"time"
)

//-----------------------------------------------------------
func TestCreateScene(t *testing.T) {
	ast := assert.New(t)

	const properName = "npc_test"
	const lenLT1Name = ""
	const lenGT40Name = "123456789123456789132465798123456798132456798123456798"
	const properRepeatType = RepeatTypeAllDay
	const lT1RepeatType = 0
	const gT3RepeatType = 4
	const properRepeatDate = "123"
	const lenLT1RepeatDate = ""
	const lenGT7RepeatDate = "123456789"
	const repeatStrRepeatDate = "11"

	tests := []struct {
		scene       Scene
		expectedRes bool
	}{
		{
			scene: Scene{
				Name:           "getOn",
				CreatorID:      1,
				AutoRun:        true,
				IsOn:           true,
				TimePeriodType: TimePeriodTypeAllDay,
				RepeatType:     RepeatTypeWorkDay,
				RepeatDate:     "12345",
				SceneConditions: []SceneCondition{
					{
						ConditionType: ConditionTypeTiming,
						TimingAt:      time.Date(0, 0, 0, 8, 30, 0, 0, time.Local),
					},
				},
				SceneTasks: []SceneTask{
					{
						Type: TaskTypeSmartDevice,
						SceneTaskDevices: []SceneTaskDevice{
							{
								DeviceID:  1,
								Action:    "switch",
								Attribute: "power",
								ActionVal: "on",
							},
						},
					},
				},
			},
			expectedRes: true,
		},
		{
			scene: Scene{
				Name:           "sleep",
				CreatorID:      1,
				AutoRun:        true,
				IsOn:           true,
				TimePeriodType: TimePeriodTypeAllDay,
				RepeatType:     RepeatTypeWorkDay,
				RepeatDate:     "12345",
				SceneConditions: []SceneCondition{
					{
						ConditionType: ConditionTypeTiming,
						TimingAt:      time.Date(0, 0, 0, 8, 30, 0, 0, time.Local),
					},
				},
				SceneTasks: []SceneTask{
					{
						Type: TaskTypeSmartDevice,
						SceneTaskDevices: []SceneTaskDevice{
							{
								DeviceID:  1,
								Action:    "switch",
								Attribute: "power",
								ActionVal: "off",
							},
						},
					},
				},
			},
			expectedRes: true,
		},
		{
			scene: Scene{
				Name:      "openLight",
				CreatorID: 1,
				AutoRun:   false,
				SceneTasks: []SceneTask{
					{
						Type: TaskTypeSmartDevice,
						SceneTaskDevices: []SceneTaskDevice{
							{
								DeviceID:  1,
								Action:    "switch",
								Attribute: "power",
								ActionVal: "on",
							},
						},
					},
				},
			},
			expectedRes: true,
		},
		{
			scene: Scene{
				Name: lenLT1Name,
			},
			expectedRes: false,
		},
		{
			scene: Scene{
				Name: lenGT40Name,
			},
			expectedRes: false,
		},
		{
			scene: Scene{
				Name:       properName,
				AutoRun:    true,
				RepeatType: lT1RepeatType,
				RepeatDate: properRepeatDate,
			},
			expectedRes: false,
		},
		{
			scene: Scene{
				Name:       properName,
				AutoRun:    true,
				RepeatType: gT3RepeatType,
				RepeatDate: properRepeatDate,
			},
			expectedRes: false,
		},
		{
			scene: Scene{
				Name:       properName,
				AutoRun:    true,
				RepeatType: properRepeatType,
				RepeatDate: repeatStrRepeatDate,
			},
			expectedRes: false,
		},
		{
			scene: Scene{
				Name:       properName,
				AutoRun:    true,
				RepeatType: properRepeatType,
				RepeatDate: lenLT1RepeatDate,
			},
			expectedRes: false,
		},
		{
			scene: Scene{
				Name:       properName,
				AutoRun:    true,
				RepeatType: properRepeatType,
				RepeatDate: lenGT7RepeatDate,
			},
			expectedRes: false,
		},
	}

	for i, t := range tests {
		err := CreateScene(&t.scene)
		if t.expectedRes {
			ast.NoError(err, "%v", i)
		} else {
			ast.Error(err, "%v", i)
		}
	}
}

func TestIsSceneNameExist(t *testing.T) {
	ast := assert.New(t)

	const existName = "getOn"
	const notExistName = "dsjfkldfjs"
	const notUseID = 0
	const correspondID = 1
	const notCorrespondID = 2

	tt := []struct {
		name        string
		id          int
		expectedRes bool
	}{
		{
			name:        existName,
			id:          notUseID,
			expectedRes: true,
		},
		{
			name:        existName,
			id:          correspondID,
			expectedRes: false,
		},
		{
			name:        existName,
			id:          notCorrespondID,
			expectedRes: true,
		},
		{
			name:        notExistName,
			id:          notUseID,
			expectedRes: false,
		},
	}

	for i, t := range tt {
		err := IsSceneNameExist(t.name, t.id)
		if !t.expectedRes {
			ast.NoError(err, "%v", i)
		} else {
			ast.Error(err, "%v", i)
		}
	}
}

func TestCheckSceneExitById(t *testing.T) {
	ast := assert.New(t)

	const exitID = 1
	const notExitID = 999

	err := CheckSceneExitById(exitID)
	ast.NoError(err)

	err = CheckSceneExitById(notExitID)
	ast.Error(err)
}

func TestGetSceneById(t *testing.T) {
	ast := assert.New(t)

	const exitID = 1
	const notExitID = 999

	s, err := GetSceneById(exitID)
	ast.NoError(err)
	ast.NotEmpty(s)

	s, err = GetSceneById(notExitID)
	ast.Error(err)
	ast.Empty(s)
}

func TestGetScenes(t *testing.T) {
	ast := assert.New(t)

	const count = 3

	scenes, err := GetScenes()
	ast.NoError(err)
	ast.Equal(count, len(scenes))
}

func TestGetSceneInfoById(t *testing.T) {
	scene, err := GetSceneInfoById(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, scene)

	scene, err = GetSceneInfoById(999)
	assert.Error(t, err)
	assert.Empty(t, scene)
}

func TestGetSceneByIDWithUnscoped(t *testing.T) {
	scene, err := GetSceneByIDWithUnscoped(1)
	assert.NoError(t, err)
	assert.NotEmpty(t, scene)

	scene, err = GetSceneByIDWithUnscoped(999)
	assert.Error(t, err)
	assert.Empty(t, scene)
}

func TestSwitchAutoSceneByID(t *testing.T) {
	ast := assert.New(t)

	const exitID = 1
	const notExitID = 999
	const on = true
	const off = false

	tt := []struct {
		id          int
		isExecute   bool
		expectedRes bool
	}{
		{
			id:          exitID,
			isExecute:   off,
			expectedRes: true,
		},
		{
			id:          exitID,
			isExecute:   on,
			expectedRes: true,
		},
		{
			id:          notExitID,
			isExecute:   off,
			expectedRes: false,
		},
	}

	for _, t := range tt {
		err := SwitchAutoSceneByID(t.id, t.isExecute)
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}

	}
}

//-----------------------------------------------------------
func TestCreateSceneTask(t *testing.T) {
	ast := assert.New(t)

	scene := Scene{
		ID:      4,
		Name:    "testSceneTask",
		AutoRun: false,
	}
	GetDB().Create(&scene)

	tt := [][]SceneTask{
		{
			{
				SceneID:      4,
				DelaySeconds: 1,
			},
		},
		{
			{
				SceneID:      4,
				DelaySeconds: 2,
			},
		},
	}

	for _, t := range tt {
		err := CreateSceneTask(t)
		ast.NoError(err, "create scene task error: %v", err)
	}
}

func TestGetSceneTasksBySceneID(t *testing.T) {
	const exitSceneID = 4
	const notExitSceneID = 999

	sceneTasks, err := GetSceneTasksBySceneID(exitSceneID)
	assert.NoError(t, err)
	assert.NotEmpty(t, sceneTasks)

	sceneTasks, err = GetSceneTasksBySceneID(notExitSceneID)
	assert.NoError(t, err)
	assert.Empty(t, sceneTasks)
}

func TestSceneTask_CheckTaskDevice(t *testing.T) {
	ast := assert.New(t)

	const testID = 555
	target := fmt.Sprintf("device-%v", strconv.Itoa(testID))

	user := User{
		ID: testID,
	}
	_ = CreateUser(&user)

	role := Role{
		ID: testID,
	}
	GetDB().Create(&role)

	var urs = []UserRole{
		{UserID: testID, RoleID: testID},
	}
	_ = CreateUserRole(urs)

	permission := struct {
		name      string
		action    string
		target    string
		attribute string
	}{"test", "control", target, "power"}
	_ = role.AddPermissionForRole(permission.name, permission.action, permission.target, permission.attribute)

	tt := []struct {
		sceneTask   SceneTask
		expectedRes bool
	}{
		{
			sceneTask: SceneTask{
				Type:             TaskTypeSmartDevice,
				SceneTaskDevices: []SceneTaskDevice{},
			},
			expectedRes: false,
		},
		{
			sceneTask: SceneTask{
				Type: TaskTypeSmartDevice,
				SceneTaskDevices: []SceneTaskDevice{
					{
						DeviceID:  1,
						Attribute: "power",
					},
				},
			},
			expectedRes: false,
		},
		{
			sceneTask: SceneTask{
				Type: TaskTypeSmartDevice,
				SceneTaskDevices: []SceneTaskDevice{
					{
						DeviceID:  testID,
						Attribute: "power",
						Action:    types.ActionSwitch,
					},
				},
			},
			expectedRes: true,
		},
	}

	for _, t := range tt {
		err := t.sceneTask.CheckTaskDevice(testID)
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}
	}

	_ = DeleteRole(testID)
	_ = DelUser(testID)
	_ = DelUserRoleByUid(testID, GetDB())
}

func TestSceneTask_CheckTaskType(t *testing.T) {
	ast := assert.New(t)

	const lt1TaskType = 0
	const gt4TaskType = 5
	const properTaskType = 2

	tt := []struct {
		sceneTask   SceneTask
		expectedRes bool
	}{
		{
			sceneTask: SceneTask{
				Type: lt1TaskType,
			},
			expectedRes: false,
		},
		{
			sceneTask: SceneTask{
				Type: gt4TaskType,
			},
			expectedRes: false,
		},
		{
			sceneTask: SceneTask{
				Type: properTaskType,
			},
			expectedRes: true,
		},
	}

	for _, t := range tt {
		err := t.sceneTask.CheckTaskType()
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}
	}
}

//-----------------------------------------------------------
func TestCreateTaskPerformItem(t *testing.T) {
	ast := assert.New(t)

	sceneTask := []SceneTask{
		{
			ID:             200,
			SceneID:        1,
			ControlSceneID: 1,
			Type:           TaskTypeSmartDevice,
		},
	}
	GetDB().Create(&sceneTask)

	tt := [][]SceneTaskDevice{
		{
			{
				SceneTaskID: 200,
				DeviceID:    1,
				Action:      "test",
				Attribute:   "test",
				ActionVal:   "0",
			},
		},
		{
			{
				SceneTaskID: 200,
				DeviceID:    2,
				Action:      "test",
				Attribute:   "test",
				ActionVal:   "0",
			},
		},
	}

	for _, t := range tt {
		err := CreateTaskPerformItem(t)
		ast.NoError(err, "create task perform item error: %v", err)
	}
}

func TestGetSceneTaskItemsByTaskID(t *testing.T) {
	ast := assert.New(t)

	const exitTaskID = 200
	const notExitTaskID = 999

	taskItems, err := GetSceneTaskItemsByTaskID(exitTaskID)
	ast.NoError(err, "get scene task items by taskid error: %v", err)
	ast.NotEmpty(taskItems)

	taskItems, err = GetSceneTaskItemsByTaskID(notExitTaskID)
	ast.NoError(err, "get scene task items by taskid error: %v", err)
	ast.Empty(taskItems)
}

func TestSceneTaskDevice_CheckPerformItem(t *testing.T) {
	ast := assert.New(t)

	tt := []struct {
		sceneTaskDevice SceneTaskDevice
		expectedRes     bool
	}{
		{
			sceneTaskDevice: SceneTaskDevice{
				DeviceID: 0,
			},
			expectedRes: false,
		},
		{
			sceneTaskDevice: SceneTaskDevice{
				DeviceID: 1,
				Action:   "",
			},
			expectedRes: false,
		},
		{
			sceneTaskDevice: SceneTaskDevice{
				DeviceID: 1,
				Action:   "test",
			},
			expectedRes: false,
		},
		{
			sceneTaskDevice: SceneTaskDevice{
				DeviceID: 1,
				Action:   types.ActionSwitch,
			},
			expectedRes: true,
		},
	}

	for _, t := range tt {
		err := t.sceneTaskDevice.checkPerformItem()
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}
	}
}

//-----------------------------------------------------------
func TestGetConditionsBySceneID(t *testing.T) {
	ast := assert.New(t)

	cs, err := GetConditionsBySceneID(1)
	ast.NoError(err)
	ast.NotEmpty(cs)

	cs, err = GetConditionsBySceneID(999)
	ast.NoError(err)
	ast.Empty(cs)
}

func TestCheckOperation(t *testing.T) {
	ast := assert.New(t)

	err := checkOperation("")
	ast.Error(err)

	err = checkOperation(types.ActionSwitch)
	ast.NoError(err)

	err = checkOperation(types.ActionSetBright)
	ast.NoError(err)

	err = checkOperation(types.ActionOnVal)
	ast.Error(err)
}

func TestConditionInfo_CheckCondition(t *testing.T) {
	ast := assert.New(t)

	const testID = 555
	target := fmt.Sprintf("device-%v", strconv.Itoa(testID))

	user := User{
		ID: testID,
	}
	_ = CreateUser(&user)

	role := Role{
		ID: testID,
	}
	GetDB().Create(&role)

	var urs = []UserRole{
		{UserID: testID, RoleID: testID},
	}
	_ = CreateUserRole(urs)

	permission := struct {
		name      string
		action    string
		target    string
		attribute string
	}{"test", "control", target, "power"}
	_ = role.AddPermissionForRole(permission.name, permission.action, permission.target, permission.attribute)

	tt := []struct {
		conditionInfo ConditionInfo
		expectedRes   bool
	}{
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					ConditionType: ConditionTypeTiming,
					DeviceID:      0,
					ConditionItem: ConditionItem{
						Operator:  "",
						Action:    "",
						ActionVal: "",
					},
				},
				Timing: 123,
			},
			expectedRes: true,
		},
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					ConditionType: ConditionTypeDeviceStatus,
					DeviceID:      testID,
					ConditionItem: ConditionItem{
						Attribute: "power",
						Action:    types.ActionSwitch,
					},
				},
				Timing: 0,
			},
			expectedRes: true,
		},
	}

	for _, t := range tt {
		err := t.conditionInfo.CheckCondition(testID)
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}
	}

	_ = DeleteRole(testID)
	_ = DelUser(testID)
	_ = DelUserRoleByUid(testID, GetDB())
}

func TestConditionInfo_checkConditionType(t *testing.T) {
	ast := assert.New(t)

	tt := []struct {
		conditionInfo ConditionInfo
		expectedRes   bool
	}{
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					ConditionType: 0,
				},
			},
			expectedRes: false,
		},
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					ConditionType: 3,
				},
			},
			expectedRes: false,
		},
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					ConditionType: 1,
				},
			},
			expectedRes: true,
		},
	}

	for _, t := range tt {
		err := t.conditionInfo.checkConditionType()
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}
	}
}

func TestConditionInfo_checkConditionTypeTiming(t *testing.T) {
	ast := assert.New(t)

	tt := []struct {
		conditionInfo ConditionInfo
		expectedRes   bool
	}{
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					DeviceID: 0,
					ConditionItem: ConditionItem{
						Operator:  "",
						Action:    "",
						ActionVal: "",
					},
				},
				Timing: 0,
			},
			expectedRes: false,
		},
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					DeviceID: 1,
					ConditionItem: ConditionItem{
						Operator:  "",
						Action:    "",
						ActionVal: "",
					},
				},
				Timing: 132456,
			},
			expectedRes: false,
		},
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					DeviceID: 0,
					ConditionItem: ConditionItem{
						Operator:  OperatorGT,
						Action:    "",
						ActionVal: "",
					},
				},
				Timing: 456456,
			},
			expectedRes: false,
		},
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					DeviceID: 0,
					ConditionItem: ConditionItem{
						Operator:  "",
						Action:    "",
						ActionVal: "",
					},
				},
				Timing: 456456,
			},
			expectedRes: true,
		},
	}

	for _, t := range tt {
		err := t.conditionInfo.checkConditionTypeTiming()
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}
	}
}

func TestConditionInfo_checkConditionDevice(t *testing.T) {
	ast := assert.New(t)

	const testID = 555
	target := fmt.Sprintf("device-%v", strconv.Itoa(testID))

	user := User{
		ID: testID,
	}
	_ = CreateUser(&user)

	role := Role{
		ID: testID,
	}
	GetDB().Create(&role)

	var urs = []UserRole{
		{UserID: testID, RoleID: testID},
	}
	_ = CreateUserRole(urs)

	permission := struct {
		name      string
		action    string
		target    string
		attribute string
	}{"test", "control", target, "power"}
	_ = role.AddPermissionForRole(permission.name, permission.action, permission.target, permission.attribute)

	tt := []struct {
		conditionInfo ConditionInfo
		expectedRes   bool
	}{
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					DeviceID: testID,
					ConditionItem: ConditionItem{
						Attribute: "power",
						Action:    types.ActionSwitch,
					},
				},
				Timing: 1,
			},
			expectedRes: false,
		},
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					DeviceID: testID,
					ConditionItem: ConditionItem{
						Attribute: "power",
						Action:    types.ActionSwitch,
					},
				},
				Timing: 0,
			},
			expectedRes: true,
		},
		{
			conditionInfo: ConditionInfo{
				SceneCondition: SceneCondition{
					DeviceID: 999,
					ConditionItem: ConditionItem{
						Attribute: "power",
						Action:    types.ActionSwitch,
					},
				},
				Timing: 0,
			},
			expectedRes: false,
		},
	}

	for _, t := range tt {
		err := t.conditionInfo.checkConditionDevice(testID)
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}
	}

	_ = DeleteRole(testID)
	_ = DelUser(testID)
	_ = DelUserRoleByUid(testID, GetDB())
}

func TestGetConditions(t *testing.T) {
	ast := assert.New(t)

	scene := Scene{
		Name:           "test1",
		CreatorID:      1,
		AutoRun:        true,
		IsOn:           true,
		TimePeriodType: TimePeriodTypeAllDay,
		RepeatType:     RepeatTypeWorkDay,
		RepeatDate:     "12345",
		SceneConditions: []SceneCondition{
			{
				ConditionType: ConditionTypeDeviceStatus,
				DeviceID:      1,
				ConditionItem: ConditionItem{
					Attribute: "power",
					Action:    types.ActionSwitch,
				},
			},
		},
		SceneTasks: []SceneTask{
			{
				Type: TaskTypeSmartDevice,
				SceneTaskDevices: []SceneTaskDevice{
					{
						DeviceID:  1,
						Action:    "switch",
						Attribute: "power",
						ActionVal: "on",
					},
				},
			},
		},
	}

	GetDB().Create(&scene)

	tt := []struct {
		deviceID    int
		attribute   string
		expectedRes bool
		isHaveScene bool
	}{
		{
			deviceID:    1,
			attribute:   "power",
			expectedRes: true,
			isHaveScene: true,
		},
		{
			deviceID:    999,
			attribute:   "power",
			expectedRes: true,
			isHaveScene: false,
		},
		{
			deviceID:    1,
			attribute:   "light",
			expectedRes: true,
			isHaveScene: true,
		},
	}

	for i, t := range tt {
		conditions, err := GetConditions(t.deviceID, t.attribute)
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}

		if t.isHaveScene {
			ast.NotEmpty(conditions)
		} else {
			ast.Empty(conditions, "%v", i)
		}
	}

	GetDB().Delete(&scene)
}

func TestGetScenesByCondition(t *testing.T) {
	ast := assert.New(t)

	scene := Scene{
		Name:           "test1",
		CreatorID:      1,
		AutoRun:        true,
		IsOn:           true,
		TimePeriodType: TimePeriodTypeAllDay,
		RepeatType:     RepeatTypeWorkDay,
		RepeatDate:     "12345",
		SceneConditions: []SceneCondition{
			{
				ConditionType: ConditionTypeDeviceStatus,
				DeviceID:      1,
				ConditionItem: ConditionItem{
					Attribute: "power",
					Action:    types.ActionSwitch,
				},
			},
		},
		SceneTasks: []SceneTask{
			{
				Type: TaskTypeSmartDevice,
				SceneTaskDevices: []SceneTaskDevice{
					{
						DeviceID:  1,
						Action:    "switch",
						Attribute: "power",
						ActionVal: "on",
					},
				},
			},
		},
	}

	GetDB().Create(&scene)

	tt := []struct {
		deviceID    int
		attribute   string
		expectedRes bool
		isHaveScene bool
	}{
		{
			deviceID:    1,
			attribute:   "power",
			expectedRes: true,
			isHaveScene: true,
		},
		{
			deviceID:    999,
			attribute:   "power",
			expectedRes: true,
			isHaveScene: false,
		},
		{
			deviceID:    1,
			attribute:   "light",
			expectedRes: true,
			isHaveScene: true,
		},
	}

	for i, t := range tt {
		scenes, err := GetScenesByCondition(t.deviceID, t.attribute)
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}

		if t.isHaveScene {
			ast.NotEmpty(scenes)
		} else {
			ast.Empty(scenes, "%v", i)
		}
	}

	GetDB().Delete(&scene)
}

//-----------------------------------------------------------
func TestConditionItem_CheckConditionItem(t *testing.T) {
	ast := assert.New(t)

	const testID = 666
	target := fmt.Sprintf("device-%v", strconv.Itoa(testID))

	user := User{
		ID: testID,
	}
	_ = CreateUser(&user)

	role := Role{
		ID: testID,
	}
	GetDB().Create(&role)

	var urs = []UserRole{
		{UserID: testID, RoleID: testID},
	}
	_ = CreateUserRole(urs)

	permission := struct {
		name      string
		action    string
		target    string
		attribute string
	}{"test", "control", target, "power"}
	_ = role.AddPermissionForRole(permission.name, permission.action, permission.target, permission.attribute)

	conditionItem := ConditionItem{
		Attribute: "power",
		Action:    types.ActionSwitch,
	}

	err := conditionItem.CheckConditionItem(testID, testID)
	ast.NoError(err)

	err = conditionItem.CheckConditionItem(3, testID)
	ast.Error(err)

	_ = DeleteRole(testID)
	_ = DelUser(testID)
	_ = DelUserRoleByUid(testID, GetDB())
}

func TestConditionItem_checkOperatorType(t *testing.T) {
	ast := assert.New(t)

	tt := []struct {
		conditionItem ConditionItem
		expectedRes   bool
	}{
		{
			conditionItem: ConditionItem{
				Operator: "",
			},
			expectedRes: true,
		},
		{
			conditionItem: ConditionItem{
				Operator: OperatorLT,
			},
			expectedRes: true,
		},
		{
			conditionItem: ConditionItem{
				Operator: "hello",
			},
			expectedRes: false,
		},
	}

	for _, t := range tt {
		err := t.conditionItem.checkOperatorType()
		if t.expectedRes {
			ast.NoError(err)
		} else {
			ast.Error(err)
		}
	}
}

func TestGetConditionItemByConditionID(t *testing.T) {
	ast := assert.New(t)

	conditionItem := ConditionItem{
		SceneConditionID: 1,
	}

	GetDB().Create(&conditionItem)
	cs, err := GetConditionItemByConditionID(1)
	ast.NoError(err)
	ast.NotEmpty(cs)

	cs, err = GetConditionItemByConditionID(999)
	ast.Error(err)
	ast.Empty(cs)

	GetDB().Delete(&conditionItem)
}

func TestDeleteScene(t *testing.T) {
	err := DeleteScene(3)
	assert.NoError(t, err)

	err = DeleteScene(999)
	assert.Error(t, err)
}
