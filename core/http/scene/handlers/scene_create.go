package handlers

import (
	"gitlab.yctc.tech/root/smartassistent.git/core/smq"
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

type CreateSceneReq struct {
	SceneInfo
}

type SceneInfo struct {
	orm.Scene
	SceneConditions []orm.ConditionInfo `json:"scene_conditions"`
	CreateTime      int64               `json:"create_time"`
	EffectStartTime int64               `json:"effect_start_time"`
	EffectEndTime   int64               `json:"effect_end_time"`
}

// CreateScene 创建场景 TODO
func CreateScene(c *gin.Context) {
	var (
		req CreateSceneReq
		err error
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		return
	}

	if err = req.check(c); err != nil {
		return
	}
	if err = req.createScene(c); err != nil {
		return
	}

	// 手动场景创建时不排进任务队列
	if req.AutoRun {
		smq.AddSceneTask(req.Scene)
	}

}

func (req *CreateSceneReq) check(c *gin.Context) (err error) {
	if !orm.JudgePermit(session.Get(c).UserID, permission.SceneAdd) {
		err = errors.New(errors.SceneCreateDeny)
		return
	}

	if err = req.validate(c); err != nil {
		return
	}
	return
}

func (req *CreateSceneReq) validate(c *gin.Context) (err error) {

	if len(req.SceneTasks) == 0 {
		err = errors.New(errors.BadRequest)
		return
	}

	if err = orm.IsSceneNameExist(req.Name, req.ID); err != nil {
		return
	}

	// 手动执行
	if !req.AutoRun {
		if req.TimePeriodType != 0 && req.ConditionLogic != 0 && req.RepeatType != 0 &&
			req.EffectStartTime != 0 && req.EffectEndTime != 0 && req.RepeatDate != "" && len(req.SceneConditions) != 0 {
			err = errors.New(errors.BadRequest)
			return
		}
	} else {
		// 自动执行
		if len(req.SceneConditions) == 0 {
			err = errors.Newf(errors.ParamIncorrectErr, "触发条件")
			return
		}
		// 生效时间不可为空
		if req.EffectStartTime == 0 || req.EffectEndTime == 0 {
			err = errors.Newf(errors.ParamIncorrectErr, "生效时间")
			return
		}

		// ConditionLogic 校验
		if req.CheckConditionLogic() {
			err = errors.Newf(errors.ParamIncorrectErr, "满足条件类型")
			return
		}
		// 生效时间类型检验
		if err = req.CheckPeriodType(); err != nil {
			return
		}

		// SceneCondition 触发条件检验
		var count int
		for _, sc := range req.SceneConditions {
			// 触发条件为满足全部时，定时触发条件只允许一个
			if sc.ConditionType == orm.ConditionTypeTiming && req.ConditionLogic == orm.MatchAllCondition {
				count++
				if count > 1 {
					err = errors.New(errors.ConditionTimingCountErr)
					return
				}

			}
			if err = sc.CheckCondition(session.Get(c).UserID); err != nil {
				return
			}
		}
	}
	// 执行任务的校验
	for _, sceneTask := range req.SceneTasks {
		if err = CheckSceneTasks(session.Get(c).UserID, sceneTask); err != nil {
			return
		}
	}
	return
}

// checkSceneTasks 执行任务校验
func CheckSceneTasks(userId int, task orm.SceneTask) (err error) {
	if err = task.CheckTaskType(); err != nil {
		err = errors.New(errors.TaskTypeErr)
		return
	}
	// 控制设备
	if task.Type == orm.TaskTypeSmartDevice {
		if err = task.CheckTaskDevice(userId); err != nil {
			return
		}
	} else {
		if err = checkTaskScene(userId, task.ControlSceneID); err != nil {
			return
		}
	}
	return
}

// checkTaskScene 校验场景任务类型
func checkTaskScene(userId, controlSceneId int) (err error) {
	// 控制场景
	if err = orm.CheckSceneExitById(controlSceneId); err != nil {
		return
	}

	var controlPermission bool
	if controlPermission, err = CheckControlPermission(controlSceneId, userId); err != nil {
		return
	}
	if !controlPermission {
		err = errors.New(errors.DeviceOrSceneControlDeny)
		return
	}

	return
}

func (req *CreateSceneReq) createScene(c *gin.Context) (err error) {
	req.Scene.CreatorID = session.Get(c).UserID
	req.Scene.EffectStart = time.Unix(req.EffectStartTime, 0)
	req.Scene.EffectEnd = time.Unix(req.EffectEndTime, 0)
	// 自动场景
	if req.AutoRun {
		req.Scene.SceneConditions = getConditionReq(req.SceneConditions)
		req.Scene.IsOn = true
	}

	req.Scene.SceneTasks = req.SceneTasks

	if err = orm.CreateScene(&req.Scene); err != nil {
		return
	}
	return
}

// getConditionReq 获取触发条件参数值
func getConditionReq(conditions []orm.ConditionInfo) (sceneConditions []orm.SceneCondition) {
	for _, condition := range conditions {
		condition.TimingAt = time.Unix(condition.Timing, 0)
		sceneConditions = append(sceneConditions, condition.SceneCondition)
	}
	return
}
