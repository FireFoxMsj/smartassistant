package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type backupDelReq struct {
	FileName string `json:"file_name"`
}

// DeleteBackup 备份列表
func DeleteBackup(c *gin.Context) {
	var (
		req backupDelReq
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
	err = supervisor.GetManager().DeleteBackup(req.FileName)
}
