package user

import (
	"github.com/sirupsen/logrus"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/setting"
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

func validate(authToken string) (err error) {

	// 校验jwt Token
	_, err = setting.ValidateAuthTokenJwt(authToken)
	return
}

func updateToken(userID int) (resp GetTokenResp, err error) {
	token := hash.GetSaToken()
	if err = entity.EditUser(userID, entity.User{
		Token: token,
	}); err != nil {
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
	authToken := c.GetHeader(AuthToken)
	logrus.Info("areaToken in request Header: ", authToken, "areaToken in sa: ", setting.GetUserCredentialAuthToken())
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

	if err = validate(authToken); err != nil {
		err = errors.Wrap(err, status.GetUserTokenAuthDeny)
		return
	}

	resp, err = updateToken(uID)
}
