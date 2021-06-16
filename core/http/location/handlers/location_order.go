package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gorm.io/gorm"
)

type locationOrderReq struct {
	LocationsOrder []int `json:"locations_id"`
}

// TODO: 批量修改房间信息, 目前只是修改房间排序
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

	if count, err = orm.GetLocationCount(); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	} else if len(req.LocationsOrder) != int(count) {
		err = errors.New(errors.BadRequest)
		return
	}

	// location_id不存在,回滚数据

	if err = orm.GetDB().Transaction(func(tx *gorm.DB) error {
		for i, locationId := range req.LocationsOrder {
			if !orm.IsLocationExist(locationId) {
				err = errors.Wrap(err, errors.LocationNotExit)
				return err
			}
			if err = orm.EditLocationSort(locationId, i+1); err != nil {
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
