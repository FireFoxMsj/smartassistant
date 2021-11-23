package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth/generate"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gopkg.in/oauth2.v3"
	"strconv"
)

type RefreshTokenReq struct {
	ClientID     string `json:"client_id"`
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

func RefreshToken(c *gin.Context) {
	var (
		req  RefreshTokenReq
		resp TokenInfo
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindJSON(&req); err != nil {
		return
	}

	tgr := &oauth2.TokenGenerateRequest{
		Request:  c.Request,
		ClientID: req.ClientID,
		Refresh:  req.RefreshToken,
	}

	clientInfo, err := entity.GetClientByClientID(req.ClientID)
	if err != nil {
		return
	}
	tgr.ClientSecret = clientInfo.ClientSecret
	claims, err := generate.DecodeJwt(req.RefreshToken)
	if err != nil {
		err = errors.New(errors.InternalServerErr)
		return
	}

	tgr.Request.Header.Set(types.AreaID, strconv.FormatUint(claims.AreaID, 10))
	ti, err := oauth.GetOauthServer().GetAccessToken(oauth2.GrantType(req.GrantType), tgr)
	if err != nil {
		logger.Errorf("refresh token failed： %v", err)
		err = errors.Wrap(err, status.ErrInvalidRefreshToken)

		return
	}

	resp = TokenInfo{
		AccessToken:     ti.GetAccess(),
		AccessTokenExp:  int64(ti.GetAccessExpiresIn().Seconds()),
		RefreshToken:    ti.GetRefresh(),
		RefreshTokenExp: int64(ti.GetRefreshExpiresIn().Seconds()),
	}
}
