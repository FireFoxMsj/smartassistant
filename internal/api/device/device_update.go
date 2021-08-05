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

// UpdateDeviceReq 修改设备接口请求参数
type UpdateDeviceReq struct {
	Name       *string `json:"name"`
	LocationID int     `json:"location_id"`
}

func (req *UpdateDeviceReq) Validate() (updateDevice entity.Device, err error) {
	if req.LocationID != 0 {
		if _, err = entity.GetLocationByID(req.LocationID); err != nil {
			return
		}
	}
	updateDevice.LocationID = req.LocationID

	if req.Name != nil {
		if err = checkDeviceName(*req.Name); err != nil {
			return
		} else {
			updateDevice.Name = *req.Name
		}
	}
	return
}

// UpdateDevice 用于处理修改设备接口的请求
func UpdateDevice(c *gin.Context) {
	var (
		err          error
		req          UpdateDeviceReq
		id           int
		updateDevice entity.Device
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

	if _, err = entity.GetDeviceByID(id); err != nil {
		err = errors.New(status.DeviceNotExist)
		return
	}

	p := types.NewDeviceUpdate(id)
	if !isPermit(c, p) {
		err = errors.Wrap(err, status.Deny)
		return
	}
	if updateDevice, err = req.Validate(); err != nil {
		return
	}

	if req.LocationID == 0 {
		// 未勾选房间, 设备与房间解绑
		if err = entity.UnBindLocationDevice(id); err != nil {
			return
		}
	}

	if err = entity.UpdateDevice(id, updateDevice); err != nil {
		return
	}

	return
}
