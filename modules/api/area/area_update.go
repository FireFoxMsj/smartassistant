package area

import (
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// UpdateAreaReq 修改家庭接口请求参数
type UpdateAreaReq struct {
	Name string `json:"name"`
}

func (req *UpdateAreaReq) Validate() (err error) {
	if err = checkAreaName(req.Name); err != nil {
		return
	}
	return
}

// UpdateArea 用于处理修改家庭接口的请求
func UpdateArea(c *gin.Context) {
	var (
		err    error
		req    UpdateAreaReq
		areaID uint64
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	areaID, err = strconv.ParseUint(c.Param("id"), 10, 64)
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

	if _, err = entity.GetAreaByID(areaID); err != nil {
		return
	}

	if err = entity.UpdateArea(areaID, req.Name); err != nil {
		return
	}
	cloud.UpdateAreaName(areaID, req.Name)
	return
}
