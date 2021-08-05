package task

import (
	"github.com/zhiting-tech/smartassistant/internal/entity"
)

// IsSceneHaveTimeCondition 场景是否有定时条件
func IsSceneHaveTimeCondition(scene entity.Scene) bool {
	for _, c := range scene.SceneConditions {
		if c.ConditionType == entity.ConditionTypeTiming {
			return true
		}
	}
	return false
}
