package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/core/smq"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

type ExecuteSceneReq struct {
	IsExecute bool `json:"is_execute"`
}

func ExecuteScene(c *gin.Context) {
	var (
		req ExecuteSceneReq
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	sceneId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	if err = orm.CheckSceneExitById(sceneId); err != nil {
		return
	}

	if err = req.executeScene(sceneId, session.Get(c).UserID); err != nil {
		return
	}
}

func (req ExecuteSceneReq) executeScene(sceneId int, userId int) (err error) {

	controlPermission, err := CheckControlPermission(sceneId, userId)
	if !controlPermission {
		err = errors.New(errors.DeviceOrSceneControlDeny)
		return
	}

	scene, err := orm.GetSceneInfoById(sceneId)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	if scene.AutoRun {
		// 自动场景先更新数据库
		if err = orm.SwitchAutoScene(&scene, req.IsExecute); err != nil {
			return
		}
	}

	if req.IsExecute {
		smq.AddSceneTask(scene)
	} else {
		smq.DelSceneTask(scene.ID)
	}

	return
}
