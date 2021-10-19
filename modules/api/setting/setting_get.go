package setting

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type GetSettingResp struct {
	UserCredentialFoundSetting entity.UserCredentialFoundSetting `json:"user_credential_found_setting"`
}

// GetSetting 获取全局配置
func GetSetting(c *gin.Context) {

	var (
		resp GetSettingResp
		err  error
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	setting := entity.GetDefaultUserCredentialFoundSetting()
	err = entity.GetSetting(entity.UserCredentialFoundType, &setting, session.Get(c).AreaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.UserCredentialFoundSetting = setting
}
