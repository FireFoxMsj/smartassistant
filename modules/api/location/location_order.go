package location

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

// locationOrderReq 调整房间列表顺序接口请求体
type locationOrderReq struct {
	LocationsOrder []int `json:"locations_id"`
}

// LocationOrder 用于处理调整房间列表顺序的请求
func LocationOrder(c *gin.Context) {
	var (
		err   error
		req   locationOrderReq
		count int64
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	u := session.Get(c)
	if count, err = entity.GetLocationCount(u.AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	} else if len(req.LocationsOrder) != int(count) {
		err = errors.New(errors.BadRequest)
		return
	}

	// location_id不存在,回滚数据
	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {
		for i, locationId := range req.LocationsOrder {
			if !entity.IsLocationExist(u.AreaID, locationId) {
				err = errors.Wrap(err, status.LocationNotExit)
				return err
			}
			if err = entity.EditLocationSort(locationId, i+1); err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
				return err
			}
		}
		return nil
	}); err != nil {
		return
	}

	return
}
