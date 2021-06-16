package handlers

import (
	"github.com/gin-gonic/gin"
	utils2 "gitlab.yctc.tech/root/smartassistent.git/core/http/area/utils"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"strconv"
)

type EditNameReq struct {
	Name string `json:"name"`
}

func (req *EditNameReq) Validate() (err error) {
	if err = utils2.CheckAreaName(req.Name); err != nil {
		return
	}
	return
}

func EditAreaName(c *gin.Context) {
	var (
		err    error
		req    EditNameReq
		areaId int
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	areaId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = req.Validate(); err != nil {
		return
	}

	if _, err = orm.GetAreaByID(areaId); err != nil {
		return
	}

	err = orm.EditAreaName(areaId, req.Name)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}
