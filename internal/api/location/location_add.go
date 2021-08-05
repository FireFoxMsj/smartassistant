package location

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"time"
)

// locationAddReq 添加房间接口请求参数
type locationAddReq struct {
	Name string `json:"name"`
}

func (req *locationAddReq) Validate() (location entity.Location, err error) {
	if err = checkLocationName(req.Name); err != nil {
		return
	} else {
		location.Name = req.Name
	}
	return
}

// AddLocation 用于处理添加房间接口的请求
func AddLocation(c *gin.Context) {
	var (
		newLocation entity.Location
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

	if err = entity.CreateLocation(&newLocation); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}
