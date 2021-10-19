package location

import (
	errors2 "errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// DelLocation 用于处理删除房间接口的请求
func DelLocation(c *gin.Context) {
	var (
		id  int
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	id, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = entity.DelLocation(id); err != nil {
		return
	}
	// 将绑定在该房间下的所有设备与该房间解绑
	if err = entity.UnBindLocationDevices(id); err != nil {
		if !errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.InternalServerErr)
		} else {
			err = nil
		}
		return
	}
	return

}
