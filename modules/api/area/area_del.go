package area

import (
	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/clouddisk"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type DelAreaReq struct {
	IsDelCloudDisk *bool `json:"is_del_cloud_disk"`
}

// DelArea 用于处理删除家庭接口的请求
func DelArea(c *gin.Context) {
	var (
		id  uint64
		err error
		req DelAreaReq
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	id, err = strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	// 校验AreaID
	if _, err = entity.GetAreaByID(id); err != nil {
		return
	}

	if req.IsDelCloudDisk != nil && *req.IsDelCloudDisk {
		// FIXME 云端没有网盘
		clouddisk.DelAreaCloudDisk(c, id)
	}

	if err = entity.DelAreaByID(id); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	cloud.RemoveSA(id)
	return
}
