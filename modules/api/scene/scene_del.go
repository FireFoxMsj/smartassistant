package scene

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/task"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// DeleteScene 用于处理删除场景接口的请求
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

	if !entity.JudgePermit(session.Get(c).UserID, types.SceneDel) {
		err = errors.New(status.SceneDeleteDeny)
		return
	}

	if err = entity.DeleteScene(sceneId); err != nil {
		return
	}

	task.GetManager().DeleteSceneTask(sceneId)

}
