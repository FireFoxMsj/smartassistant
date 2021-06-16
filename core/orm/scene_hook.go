package orm

import (
	"gorm.io/gorm"
	"strings"
)

const (
	sceneNameMinLength = 1
	sceneNameMaxLength = 40
)

func (r *Scene) BeforeSave(tx *gorm.DB) (err error) {
	if err = r.CheckRepeatConfig(); err != nil {
		return
	}

	r.RepeatDate = strings.TrimSpace(r.RepeatDate)

	if err = r.CheckSceneNameLength(); err != nil {
		return
	}

	return nil
}
