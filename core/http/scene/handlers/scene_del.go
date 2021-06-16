package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/core/smq"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
	"strconv"
)

func DeleteScene(c *gin.Context) {
	var err error
	defer func() {
		response.HandleResponse(c, err, nil)

	}()

	sceneId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	if !orm.JudgePermit(session.Get(c).UserID, permission.SceneDel) {
		err = errors.New(errors.SceneDeleteDeny)
		return
	}

	if err = orm.DeleteScene(sceneId); err != nil {
		return
	}

	smq.DelSceneTask(sceneId)

}
