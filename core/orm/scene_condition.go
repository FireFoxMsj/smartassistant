package orm

import (
	errors2 "errors"
	"time"

	"gorm.io/gorm"

	utils2 "gitlab.yctc.tech/root/smartassistent.git/utils"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

type ConditionType int

const (
	ConditionTypeTiming ConditionType = iota + 1
	ConditionTypeDeviceStatus
)

type OperatorType string

const (
	OperatorGT OperatorType = ">"
	OperatorLT OperatorType = "<"
	OperatorEQ OperatorType = "="
)

// SceneCondition 场景条件
type SceneCondition struct {
	ID            int           `json:"id"`
	SceneID       int           `json:"scene_id"`
	ConditionType ConditionType `json:"condition_type"`
	TimingAt      time.Time     `json:"-"` // 定时在某个时间

	// 设备有关配置
	DeviceID      int           `json:"device_id"` // 或某个设备状态变化时
	ConditionItem ConditionItem `json:"condition_item" gorm:"constraint:OnDelete:CASCADE;"`
}

func (d SceneCondition) TableName() string {
	return "scene_conditions"
}

func GetConditionsBySceneID(sceneID int) (conditions []SceneCondition, err error) {
	err = GetDB().Where("scene_id = ?", sceneID).Find(&conditions).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)

	}
	return
}

// ConditionItem 记录设备各个功能点的条件具体配置
type ConditionItem struct {
	ID               int          `json:"id"`
	SceneConditionID int          `json:"scene_condition_id"` //
	Action           string       `json:"action"`             // 功能点，比如开关、亮度、色温
	Attribute        string       `json:"attribute"`          // 功能对应的属性，比如开关电源，开关左键
	Operator         OperatorType `json:"operator"`           // 操作符，大于、小于、等于
	ActionVal        string       `json:"action_val"`         //
}

func (d ConditionItem) TableName() string {
	return "scene_condition_items"
}

func GetConditionItemByConditionID(conditionID int) (conditionItem ConditionItem, err error) {
	err = GetDB().Where("scene_condition_id = ?", conditionID).First(&conditionItem).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.SceneConditionNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

type ConditionInfo struct {
	SceneCondition
	Timing int64 `json:"timing"`
}

// CheckSceneConditions 触发条件校验
func (c ConditionInfo) CheckCondition(userId int) (err error) {
	if err = c.checkConditionType(); err != nil {
		return
	}

	// 定时类型
	if c.ConditionType == ConditionTypeTiming {
		if err = c.checkConditionTypeTiming(); err != nil {
			return
		}
	} else {
		// 设备状态变化时
		if err = c.checkConditionDevice(userId); err != nil {
			return
		}
	}
	return
}

// checkConditionType 校验触发条件类型
func (c ConditionInfo) checkConditionType() (err error) {
	if c.ConditionType < ConditionTypeTiming || c.ConditionType > ConditionTypeDeviceStatus {
		err = errors.Newf(errors.ParamIncorrectErr, "触发条件类型")
		return
	}
	return
}

// checkConditionTypeTiming 校验定时类型
func (c ConditionInfo) checkConditionTypeTiming() (err error) {
	if c.Timing == 0 || c.DeviceID != 0 || c.ConditionItem.Operator != "" || c.ConditionItem.Action != "" || c.ConditionItem.ActionVal != "" {
		err = errors.New(errors.ConditionMisMatchTypeAndConfigErr)
		return
	}
	return
}

// checkConditionDevice 校验设备类型
func (c ConditionInfo) checkConditionDevice(userId int) (err error) {
	if c.DeviceID == 0 && c.ConditionItem.Action == "" || c.Timing != 0 {
		err = errors.New(errors.ConditionMisMatchTypeAndConfigErr)
		return
	}
	// 设备对应的功能点条件配置

	if err = c.ConditionItem.CheckConditionItem(userId, c.DeviceID); err != nil {
		return
	}
	return
}

// CheckConditionItem 触发条件为设备状态变化时，校验对应参数
func (item ConditionItem) CheckConditionItem(userId, deviceId int) (err error) {
	if err = item.checkOperatorType(); err != nil {
		return
	}

	// 设备控制权限的判断
	if !IsDeviceControlPermit(userId, deviceId, item.Attribute) {
		err = errors.New(errors.DeviceOrSceneControlDeny)
		return
	}

	if err = checkOperation(item.Action); err != nil {
		return

	}
	return
}

// checkOperatorType() 校验操作类型
func (item ConditionItem) checkOperatorType() (err error) {
	var opMap = map[OperatorType]bool{
		OperatorGT: true,
		OperatorLT: true,
		OperatorEQ: true,
	}

	if item.Operator != "" {
		if _, ok := opMap[item.Operator]; !ok {
			err = errors.Newf(errors.ParamIncorrectErr, "设备操作符")
			return
		}
	}
	return
}

// CheckOperation 设备action校验
func checkOperation(action string) (err error) {
	if action == "" {
		return
	}
	if _, ok := utils2.ActionMap[action]; !ok {
		err = errors.New(errors.DeviceActionErr)
		return
	}
	return
}

// GetScenesByCondition 根据条件获取场景
func GetScenesByCondition(deviceID int, attr string) (scenes []Scene, err error) {
	conds, err := GetConditions(deviceID, attr)
	var sceneIDs []int
	for _, cond := range conds {
		sceneIDs = append(sceneIDs, cond.SceneID)
	}
	if len(sceneIDs) == 0 {
		return
	}
	if err = GetDB().Where("auto_run = true and id in (?)", sceneIDs).Find(&scenes).Error; err != nil {
		return
	}

	return
}

// GetConditions 获取符合设备属性的条件
func GetConditions(deviceID int, attr string) (conds []SceneCondition, err error) {
	if err = GetDB().Where("device_id=?", deviceID).
		Preload("ConditionItem", "attribute=?", attr).
		Find(&conds).Error; err != nil {
		return
	}
	return
}
