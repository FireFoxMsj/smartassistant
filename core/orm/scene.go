package orm

import (
	errors2 "errors"
	"strconv"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	utils2 "gitlab.yctc.tech/root/smartassistent.git/utils"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
)

type TimePeriodType int

const (
	TimePeriodTypeAllDay TimePeriodType = iota + 1
	TimePeriodTypeCustom
)

type RepeatType int

const (
	RepeatTypeAllDay RepeatType = iota + 1
	RepeatTypeWorkDay
	RepeatTypeCustom
)

const (
	MatchAllCondition = 1 // 全部满足
	MatchAnyCondition = 2 // 任一满足
)

// Scene 场景
type Scene struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	ConditionLogic int    `json:"condition_logic"` // 1 为 全部满足，2为满足任一

	// 生效时间的配置
	TimePeriodType TimePeriodType `json:"time_period"` // 全天1、时间段2
	EffectStart    time.Time      `json:"-"`
	EffectEnd      time.Time      `json:"-"`

	// 重复执行的配置
	RepeatType RepeatType `json:"repeat_type"` // 每天1，工作日2，自定义3
	RepeatDate string     `json:"repeat_date"` // 自定义的情况下：1234567

	// 设置为手动：false，则不能再设置其他两种
	AutoRun bool `json:"auto_run"` // true 就需要设置scene_condition，false 只需表示手动
	// 场景会自动执行: true
	IsOn bool `json:"is_on"`

	CreatorID       int              `json:"creator_id"`
	CreatedAt       time.Time        `json:"-"`
	SceneConditions []SceneCondition `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	SceneTasks      []SceneTask      `json:"scene_tasks" gorm:"constraint:OnDelete:CASCADE;"`
	Deleted         gorm.DeletedAt   `json:"-"`
}

func (d Scene) TableName() string {
	return "scenes"
}

func GetScenes() (scenes []Scene, err error) {
	err = GetDB().Find(&scenes).Error
	return
}

func CreateScene(scene *Scene) (err error) {
	if err = GetDB().Create(scene).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

func GetSceneById(id int) (scene Scene, err error) {
	if err = GetDB().Where("id=?", id).First(&scene).Error; err != nil {
		return
	}
	return
}

// GetSceneInfoById 获取场景所有信息
func GetSceneInfoById(id int) (scene Scene, err error) {
	if err = GetDB().
		Preload("SceneConditions.ConditionItem").
		Preload("SceneTasks.SceneTaskDevices").
		First(&scene, id).Error; err != nil {
		return
	}
	return
}

func GetDeletedSceneByID(id int) (scene Scene, err error) {
	err = GetDB().Unscoped().
		Preload("SceneConditions.ConditionItem").
		Preload("SceneTasks.SceneTaskDevices").
		First(&scene, id).Error
	return
}

func IsSceneNameExist(name string, sceneId int) (err error) {
	var db *gorm.DB
	if sceneId != 0 {
		db = GetDB().Where("id != ? and name = ?", sceneId, name)
	} else {
		db = GetDB().Where("name = ?", name)
	}

	err = db.First(&Scene{}).Error
	if err != nil && errors2.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	err = errors.New(errors.SceneNameExist)
	return err
}

// CheckRepeatConfig 校验重复执行的配置
func (s Scene) CheckRepeatConfig() (err error) {
	if s.AutoRun {
		if s.RepeatType < RepeatTypeAllDay || s.RepeatType > RepeatTypeCustom {
			err = errors.Newf(errors.ParamIncorrectErr, "重复执行配置")
			return
		}

		if !utils2.CheckIllegalRepeatDate(s.RepeatDate) {
			err = errors.Newf(errors.ParamIncorrectErr, "重复生效时间")
			return
		}

	}
	return
}

func (s Scene) CheckSceneNameLength() (err error) {
	if s.Name == "" || utf8.RuneCountInString(s.Name) < sceneNameMinLength || utf8.RuneCountInString(s.Name) > sceneNameMaxLength {
		err = errors.New(errors.BadRequest)
		return
	}
	return
}

func CheckSceneExitById(sceneId int) (err error) {
	_, err = GetSceneById(sceneId)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New(errors.SceneNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}
	return
}

// CheckConditionLogic 校验满足条件
func (s Scene) CheckConditionLogic() bool {
	return s.ConditionLogic != MatchAllCondition && s.ConditionLogic != MatchAnyCondition

}

// 生效时间类型校验
func (s Scene) CheckPeriodType() (err error) {
	if s.TimePeriodType < TimePeriodTypeAllDay || s.TimePeriodType > TimePeriodTypeCustom {
		err = errors.Newf(errors.ParamIncorrectErr, "生效时间类型")
		return
	}
	return
}

func DeleteScene(sceneId int) (err error) {
	scene := Scene{ID: sceneId}
	err = GetDB().First(&scene).Delete(&scene).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.SceneNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

// SwitchAutoScene 切换自动场景开关
func SwitchAutoScene(scene *Scene, isExecute bool) error {

	updateMap := map[string]interface{}{
		"is_on": isExecute,
	}
	if err := GetDB().Model(&scene).Updates(&updateMap).Error; err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	return nil
}

// SwitchAutoSceneByID 切换自动场景开关
func SwitchAutoSceneByID(sceneID int, isExecute bool) error {

	scene, err := GetSceneInfoById(sceneID)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(errors.SceneNotExist)
		}
		return errors.Wrap(err, errors.InternalServerErr)
	}
	return SwitchAutoScene(&scene, isExecute)
}

// GetPendingScenesByTime 根据时间获取待执行的场景
func GetPendingScenesByTime(t time.Time) (scenes []Scene, err error) {
	weekDay := strconv.Itoa(int(t.Weekday()))
	if err = GetDB().Where("auto_run=? and is_on=? and repeat_date like ?", true, true, "%"+weekDay+"%").
		Preload("SceneConditions.ConditionItem").
		Preload("SceneTasks.SceneTaskDevices").
		Find(&scenes).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}
