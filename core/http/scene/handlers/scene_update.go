package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/core/smq"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"

	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type UpdateSceneReq struct {
	DelConditionIds []int `json:"del_condition_ids"`
	DelTaskIds      []int `json:"del_task_ids"`
	CreateSceneReq
}

func UpdateScene(c *gin.Context) {
	var (
		req UpdateSceneReq
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	sceneId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	if err = c.BindJSON(&req); err != nil {
		return

	}

	if err = req.validateRequest(sceneId, c); err != nil {
		return
	}

	req.wrapReq()

	if err = req.updateScene(sceneId); err != nil {
		return
	}

	if e := smq.RestartSceneTask(sceneId); e != nil {
		log.Println("restart scene task err:", e)
	}

}

func (req *UpdateSceneReq) validateRequest(sceneId int, c *gin.Context) (err error) {
	scene, err := orm.GetSceneById(sceneId)
	if err != nil {
		return
	}
	// 场景类型不允许修改
	if scene.AutoRun != req.AutoRun {
		err = errors.New(errors.SceneTypeForbidModify)
		return
	}

	if err = req.CreateSceneReq.validate(c); err != nil {
		return
	}
	return
}

func (req UpdateSceneReq) updateScene(sceneId int) (err error) {
	// TODO 后续调整为使用事务
	if err = orm.GetDB().Session(&gorm.Session{FullSaveAssociations: true}).Where("id=?", sceneId).Updates(&req.Scene).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if err = req.delConditions(sceneId); err != nil {
		return
	}

	if err = req.delTasks(sceneId); err != nil {
		return
	}

	return
}

func (req *UpdateSceneReq) wrapReq() {
	req.Scene.CreatedAt = time.Unix(req.CreateTime, 0)
	req.Scene.EffectStart = time.Unix(req.EffectStartTime, 0)
	req.Scene.EffectEnd = time.Unix(req.EffectEndTime, 0)

	for _, sc := range req.SceneConditions {
		sceneCondition := sc.SceneCondition
		sceneCondition.TimingAt = time.Unix(sc.Timing, 0)

		req.Scene.SceneConditions = append(req.Scene.SceneConditions, sceneCondition)
	}
}

// delConditions 删除触发条件
func (req *UpdateSceneReq) delConditions(sceneId int) (err error) {
	if len(req.DelConditionIds) == 0 {
		return
	}

	var conditions []orm.SceneCondition
	for _, id := range req.DelConditionIds {
		condition := orm.SceneCondition{ID: id, SceneID: sceneId}
		conditions = append(conditions, condition)
	}
	if err = orm.GetDB().Delete(conditions).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

// delTasks 删除执行任务
func (req *UpdateSceneReq) delTasks(sceneId int) (err error) {
	if len(req.DelTaskIds) == 0 {
		return
	}

	var tasks []orm.SceneTask
	for _, id := range req.DelTaskIds {
		task := orm.SceneTask{ID: id, SceneID: sceneId}
		tasks = append(tasks, task)
	}
	if err = orm.GetDB().Delete(tasks).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}
