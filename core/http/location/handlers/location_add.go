package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/location/utils"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"time"
)

type locationAddReq struct {
	Name string `json:"name"`
}

func (req *locationAddReq) Validate() (location orm.Location, err error) {
	if err = utils.CheckLocationName(req.Name); err != nil {
		return
	} else {
		location.Name = req.Name
	}
	return
}

func AddLocation(c *gin.Context) {
	var (
		newLocation orm.Location
		req         locationAddReq
		err         error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if newLocation, err = req.Validate(); err != nil {
		return
	}

	newLocation.CreatedAt = time.Now()

	if err = orm.CreateLocation(&newLocation); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}
