package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	utils2 "gitlab.yctc.tech/root/smartassistent.git/core/http/device/utils"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type UpdateDeviceReq struct {
	Name       *string `json:"name"`
	LocationID int     `json:"location_id"`
}

func (req *UpdateDeviceReq) Validate() (updateDevice orm.Device, err error) {
	if req.LocationID != 0 {
		if _, err = orm.GetLocationByID(req.LocationID); err != nil {
			return
		}
	}
	updateDevice.LocationID = req.LocationID

	if req.Name != nil {
		if err = utils2.CheckDeviceName(*req.Name); err != nil {
			return
		} else {
			updateDevice.Name = *req.Name
		}
	}
	return
}

func UpdateDevice(c *gin.Context) {
	var (
		err          error
		req          UpdateDeviceReq
		id           int
		updateDevice orm.Device
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()
	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	id, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	p := permission.NewDeviceUpdate(id)
	if !isPermit(c, p) {
		err = errors.Wrap(err, errors.Deny)
		return
	}
	if updateDevice, err = req.Validate(); err != nil {
		return
	}

	if req.LocationID == 0 {
		// 未勾选房间, 设备与房间解绑
		if err = orm.UnBindLocationDevice(id); err != nil {
			return
		}
	}

	if err = orm.UpdateDevice(id, updateDevice); err != nil {
		return
	}

	return
}
