package handlers

import (
	errors2 "errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type infoResp struct {
	Name          string `json:"name"`
	LocationCount int64  `json:"location_count"`
	RoleCount     int    `json:"role_count"`
}

type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func InfoArea(c *gin.Context) {
	var (
		err           error
		resp          infoResp
		areaId        int
		locationCount int64
		area          orm.Area
		roles         []orm.Role
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	areaId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if area, err = orm.GetAreaByID(areaId); err != nil {
		return
	}
	resp.Name = area.Name

	if locationCount, err = orm.GetLocationCount(); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			locationCount = 0
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}
	if roles, err = orm.GetRoles(); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.RoleCount = len(roles)
	resp.LocationCount = locationCount
	return

}
