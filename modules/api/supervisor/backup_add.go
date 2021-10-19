package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type backupAddReq struct {
	Name string `json:"name"`
}

// AddBackup 创建并且启动备份
func AddBackup(c *gin.Context) {
	var (
		req backupAddReq
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()
	err = c.BindJSON(&req)
	if err != nil {
		logger.Warnf("request error %v", err)
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	err = supervisor.GetManager().StartBackupJob(req.Name)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
}
