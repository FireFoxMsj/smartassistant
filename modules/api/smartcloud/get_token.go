package smartcloud

import (
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/hash"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

const (
	AuthToken = "Auth-Token"
)

type GetTokenResp struct {
	Token string `json:"token"`
}

func updateToken(userID int, areaID uint64, c *gin.Context) (resp GetTokenResp, err error) {
	key := hash.GetSaUserKey()
	var u = entity.User{Key: key, ID: userID, AreaID: areaID}
	if err = entity.EditUser(userID, u); err != nil {
		return
	}
	token, err := oauth.GetSAUserToken(u, c.Request)
	if err != nil {
		return
	}
	resp = GetTokenResp{
		Token: token,
	}
	return
}

// 获取找回用户凭证
func GetToken(c *gin.Context) {
	var (
		err  error
		resp GetTokenResp
		uID  int
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	uID, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	logrus.Info("areaToken in request Header: ", c.GetHeader(types.SATokenKey))
	userInfo, err := entity.GetUserByID(uID)
	if err != nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}
	// 获取配置
	setting := entity.GetDefaultUserCredentialFoundSetting()
	if err = entity.GetSetting(entity.UserCredentialFoundType, &setting, userInfo.AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 判断是否允许找回找回凭证
	if !setting.UserCredentialFound {
		err = errors.New(status.GetUserTokenDeny)
		return
	}
	resp, err = updateToken(uID, userInfo.AreaID, c)
}
