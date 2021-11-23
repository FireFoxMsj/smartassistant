// Package middleware GIN 框架中间件
package middleware

import (
	"context"
	"fmt"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/reverseproxy"
)

// RequireAccount 用户需要登录才可访问对应的接口
func RequireAccount(c *gin.Context) {
	if err := verifyAccessToken(c); err != nil {
		response.HandleResponse(c, err, nil)
		c.Abort()
		return
	}
}

func verifyAccessToken(c *gin.Context) (err error) {
	accessToken := c.GetHeader(types.SATokenKey)
	if accessToken == "" {
		accessToken = c.GetHeader(types.ScopeTokenKey)
		// 将token写入smart-assistant-token 头中，供session.Get()方法使用
		c.Request.Header.Set(types.SATokenKey, accessToken)
	}
	_, err = oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
	if err != nil {
		var uerr = errors.New(status.UserNotExist)
		if err.Error() == uerr.Error() {
			return uerr
		}
		err = errors.Wrap(err, status.RequireLogin)
		return err
	}
	return
}

// RequireOwner 拥有者才能访问
func RequireOwner(c *gin.Context) {
	u := session.Get(c)
	if u == nil {
		response.HandleResponse(c, nil, nil)
		c.Abort()
		return
	}
	if u.IsOwner {
		c.Next()
		return
	}
	response.HandleResponse(c, errors.New(status.Deny), nil)
	c.Abort()
	return
}

// WithScope 校验用户权限
func WithScope(scope string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		accessToken := ctx.GetHeader(types.SATokenKey)
		ti, _ := oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
		if !strings.Contains(ti.GetScope(), scope) {
			err := errors.New(status.Deny)
			response.HandleResponse(ctx, err, nil)
			ctx.Abort()
			return
		}
	}

}

// RequireToken 使用token验证身份，不依赖cookies.
func RequireToken(c *gin.Context) {

	uToken := c.Request.Header.Get(types.SATokenKey)
	queryToken := c.Query("token")
	if queryToken != "" && uToken == "" {
		c.Request.Header.Add(types.SATokenKey, queryToken)
	}

	if session.GetUserByToken(c) != nil {
		c.Next()
		return
	}
	response.HandleResponse(c, errors.New(status.RequireLogin), nil)
	c.Abort()
	return
}

func Middleware(sessionName string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		sessions.Sessions(sessionName, session.GetStore())(ctx)
	}
}

func DefaultMiddleware() func(ctx *gin.Context) {
	return Middleware(session.DefaultSessionName)
}

// RequirePermission 判断是否有权限
func RequirePermission(p types.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := session.Get(c)
		if u != nil {
			if entity.JudgePermit(u.UserID, p) {
				c.Next()
				return
			}
		}
		err := errors.New(status.Deny)
		response.HandleResponse(c, err, nil)
		c.Abort()

	}
}

// ProxyToPlugin 根据路径转发到后端插件
func ProxyToPlugin(ctx *gin.Context) {

	path := ctx.Param("plugin")
	if up, err := reverseproxy.GetManager().GetUpstream(path); err != nil {
		response.HandleResponseWithStatus(ctx, http.StatusBadGateway, err, nil)
	} else {
		req := ctx.Request.Clone(context.Background())

		user := session.Get(ctx)
		if user != nil {
			req.Header.Add("scope-user-id", strconv.Itoa(user.UserID))
		}

		// 替换插件静态文件地址（api/static/:sa_id/plugin/demo -> api/plugin/demo）
		oldPrefix := fmt.Sprintf("%s/plugin/%s", url.StaticPath(), path)
		newPrefix := fmt.Sprintf("api/plugin/%s", path)
		req.URL.Path = strings.Replace(req.URL.Path, oldPrefix, newPrefix, 1)
		logger.Printf("serve request from %s to %s", ctx.Request.URL.Path, req.URL.Path)
		up.Proxy.ServeHTTP(ctx.Writer, req)
	}
}

// ValidateSCReq 校验来自sc的请求
func ValidateSCReq(c *gin.Context) {
	accessToken := c.GetHeader("Auth-Token")
	_, err := oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
	if err != nil {
		err = errors.New(status.Deny)
		response.HandleResponse(c, err, nil)
		c.Abort()
		return
	}
}
