package setting

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type GetUserCredentialFoundReq struct {
	UserCredentialFoundSetting *entity.UserCredentialFoundSetting `json:"user_credential_found_setting"`
}

// UpdateSetting 修改全局配置
func UpdateSetting(c *gin.Context) {

	var (
		req         GetUserCredentialFoundReq
		err         error
		sessionUser *session.User
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	// 只有SA拥有者才能设置
	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	// 修改是否允许找回用户凭证的配置
	if req.UserCredentialFoundSetting != nil {
		err = req.UpdateUserCredentialFound(sessionUser.AreaID)
	}
}

func (req *GetUserCredentialFoundReq) UpdateUserCredentialFound(areaID uint64) (err error) {

	// 更新是否允许找回用户凭证
	setting := req.UserCredentialFoundSetting
	err = entity.UpdateSetting(entity.UserCredentialFoundType, &setting, areaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 发送找回用户凭证的认证token给SC
	if setting.UserCredentialFound {
		go SendUserCredentialAuthTokenToSC(areaID)
	}

	return nil
}
