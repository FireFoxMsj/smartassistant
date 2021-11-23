package scope

import (
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/server"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
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
	Scopes   []string `json:"scopes"`
	SAUserID *int     `json:"sa_user_id"` // 第三方云通过sc授权时需传对应sa用户的id
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
		if _, ok := types.Scopes[scope]; !ok {
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
		uKey string
		uID  int
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = req.validateRequest(c); err != nil {
		return
	}

	sessionUser := session.Get(c)
	uKey = sessionUser.Key
	uID = sessionUser.UserID
	if req.SAUserID != nil && *req.SAUserID != 0 {
		u, err := entity.GetUserByID(*req.SAUserID)
		if err != nil {
			return
		}
		uKey = u.Key
		uID = u.ID
	}

	expireTime := expiresIn
	if c.GetHeader(types.VerificationKey) != "" {
		expireTime = cloudExpireIn
	}

	u := session.Get(c)
	accessToken := c.GetHeader(types.SATokenKey)
	ti, _ := oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
	c.Request.Header.Set(types.AreaID, strconv.FormatUint(u.AreaID, 10))
	c.Request.Header.Set(types.UserKey, uKey)
	tgr := &server.AuthorizeRequest{
		ResponseType:   oauth2.Token,
		ClientID:       ti.GetClientID(),
		UserID:         strconv.Itoa(uID),
		Scope:          strings.Join(req.Scopes, ","),
		AccessTokenExp: expireTime,
		Request:        c.Request,
	}

	// TODO 使用oauth2生成scope_token，后续需要与前端联调去除
	tokenInfo, err := oauth.GetOauthServer().GetAuthorizeToken(tgr)
	if err != nil {
		logger.Printf("get oauth2 token error %s", err.Error())
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	resp.ScopeToken.Token = tokenInfo.GetAccess()
	resp.ScopeToken.ExpiresIn = int(expireTime / time.Second)
}
