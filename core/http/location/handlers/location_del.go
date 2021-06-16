package handlers

import (
	errors2 "errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

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

	if err = orm.DelLocation(id); err != nil {
		return
	}
	// 将绑定在该房间下的所有设备与该房间解绑
	if err = orm.UnBindLocationDevices(id); err != nil {
		if !errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.InternalServerErr)
		} else {
			err = nil
		}
		return
	}
	return

}
