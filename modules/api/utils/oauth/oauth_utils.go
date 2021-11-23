package oauth

import (
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/server"
	"net/http"
	"strconv"
)

// GetSAUserToken 获取SA用户Token,提供给添加sa设备，扫码加入使用
func GetSAUserToken(user entity.User, req *http.Request) (token string, err error) {
	saClient, _ := entity.GetSAClient()

	authReq := &server.AuthorizeRequest{
		ResponseType: oauth2.Token,
		ClientID:     saClient.ClientID,
		Scope:        saClient.AllowScope,
		UserID:       strconv.Itoa(user.ID),
		Request:      req,
	}
	authReq.Request.Header.Set(types.AreaID, strconv.FormatUint(user.AreaID, 10))
	authReq.Request.Header.Set(types.UserKey, user.Key)

	ti, err := GetOauthServer().GetAuthorizeToken(authReq)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return ti.GetAccess(), nil
}
