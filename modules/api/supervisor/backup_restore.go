package supervisor

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type restoreReq struct {
	Name string `json:"name"`
}

// Restore 启动恢复
func Restore(c *gin.Context) {
	var (
		req restoreReq
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()
	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	err = supervisor.GetManager().StartRestoreJob(req.Name)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrap(err, status.FileNotExistErr)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
}
