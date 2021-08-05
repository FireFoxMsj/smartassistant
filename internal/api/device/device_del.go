package device

import (
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// DelDevice 用于处理删除设备接口的请求
func DelDevice(c *gin.Context) {
	var (
		err      error
		deviceId int
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	deviceId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	p := types.NewDeviceDelete(deviceId)
	if !isPermit(c, p) {
		err = errors.Wrap(err, status.Deny)
		return
	}
	if err = entity.DelDeviceByID(deviceId); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return

}
