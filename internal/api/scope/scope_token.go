package scope

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/config"
	"github.com/zhiting-tech/smartassistant/internal/types"
	jwt2 "github.com/zhiting-tech/smartassistant/internal/utils/jwt"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"log"
	"strings"
	"time"
)

type token struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

type scopeTokenResp struct {
	ScopeToken token `json:"scope_token"`
}

var (
	expiresIn     = time.Hour * 24 * 30
	cloudExpireIn = expiresIn * 6 // 用于云端控制，时间稍微设长一点
)

type scopeTokenReq struct {
	Scopes []string `json:"scopes"`
}

func (req *scopeTokenReq) validateRequest(c *gin.Context) (err error) {
	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if len(req.Scopes) == 0 {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	// 必须是允许范围内的scope
	for _, scope := range req.Scopes {
		if _, ok := scopes[scope]; !ok {
			err = errors.New(errors.BadRequest)
			return
		}
	}
	return
}

// 根据用户选择，使用用户的token作为生成 JWT
func scopeToken(c *gin.Context) {
	var (
		req  scopeTokenReq
		resp scopeTokenResp
		err  error
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = req.validateRequest(c); err != nil {
		return
	}

	claims := jwt.MapClaims{
		"uid":    session.Get(c).UserID,
		"sa_id":  config.GetConf().SmartAssistant.ID,
		"scopes": strings.Join(req.Scopes, ","),
	}

	expireTime := expiresIn
	if c.GetHeader(types.VerificationKey) != "" {
		expireTime = cloudExpireIn
	}

	claims["exp"] = time.Now().Add(expireTime).Unix()

	token, err := jwt2.GenerateUserJwt(claims, session.Get(c))
	if err != nil {
		log.Printf("generate jwt error %s", err.Error())
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	resp.ScopeToken.Token = token

	resp.ScopeToken.ExpiresIn = int(expireTime / time.Second)

}
