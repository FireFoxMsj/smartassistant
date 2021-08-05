package location

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
)

// updateLocationReq 修改房间接口请求参数
type updateLocationReq struct {
	Name *string `json:"name"`
	Sort *int    `json:"sort"`
}

func (req *updateLocationReq) Validate(locationId int) (updateLocation entity.Location, err error) {
	var (
		location entity.Location
	)
	if req.Name != nil {
		if err = checkLocationName(*req.Name); err != nil {
			return
		}

		if location, err = entity.GetLocationByID(locationId); err != nil {
			return
		}

		if location.Name != *req.Name {
			if entity.LocationNameExist(*req.Name) {
				err = errors.Wrap(err, status.LocationNameExist)
				return
			}
			updateLocation.Name = *req.Name
		}

	}

	if req.Sort != nil {
		if err = checkLocationSort(*req.Sort); err != nil {
			return
		} else {
			updateLocation.Sort = *req.Sort
		}
	}
	return
}

// UpdateLocation 用于处理修改房间接口的请求
func UpdateLocation(c *gin.Context) {
	var (
		req            updateLocationReq
		locationId     int
		updateLocation entity.Location
		err            error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if locationId, err = strconv.Atoi(c.Param("id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if updateLocation, err = req.Validate(locationId); err != nil {
		return
	}

	if err = entity.UpdateLocation(locationId, updateLocation); err != nil {
		return
	}
	return

}
