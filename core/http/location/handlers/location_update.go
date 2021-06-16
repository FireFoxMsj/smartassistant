package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/location/utils"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"strconv"
)

type updateLocationReq struct {
	Name *string `json:"name"`
	Sort *int    `json:"sort"`
}

func (req *updateLocationReq) Validate(locationId int) (updateLocation orm.Location, err error) {
	var (
		location orm.Location
	)
	if req.Name != nil {
		if err = utils.CheckLocationName(*req.Name); err != nil {
			return
		}

		if location, err = orm.GetLocationByID(locationId); err != nil {
			return
		}

		if location.Name != *req.Name {
			if orm.LocationNameExist(*req.Name) {
				err = errors.Wrapf(err, errors.NameExist, "房间名")
				return
			}
			updateLocation.Name = *req.Name
		}

	}

	if req.Sort != nil {
		if err = utils.CheckLocationSort(*req.Sort); err != nil {
			return
		} else {
			updateLocation.Sort = *req.Sort
		}
	}
	return
}

func UpdateLocation(c *gin.Context) {
	var (
		req            updateLocationReq
		locationId     int
		updateLocation orm.Location
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

	if err = orm.UpdateLocation(locationId, updateLocation); err != nil {
		return
	}
	return

}
