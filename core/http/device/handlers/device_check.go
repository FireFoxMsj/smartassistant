package handlers

import (
	errors2 "errors"
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gorm.io/gorm"
)

type CheckSaDeviceResp struct {
	IsBind bool `json:"is_bind"`
}

func CheckSaDevice(c *gin.Context) {
	var (
		err  error
		resp CheckSaDeviceResp
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if _, err = orm.GetSaDevice(); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.IsBind = true
	return
}
