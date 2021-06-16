package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

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

	p := permission.NewDeviceDelete(deviceId)
	if !isPermit(c, p) {
		err = errors.Wrap(err, errors.Deny)
		return
	}
	if err = orm.DelDeviceByID(deviceId); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return

}

func isPermit(c *gin.Context, p permission.Permission) bool {
	u := session.Get(c)
	return u != nil && orm.JudgePermit(u.UserID, p)
}
