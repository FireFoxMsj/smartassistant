// Package middleware GIN 框架中间件
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/jwt"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/reverseproxy"
)

// RequireAccount 用户需要登录才可访问对应的接口
func RequireAccount(c *gin.Context) {
	u := session.Get(c)
	if u == nil {
		response.HandleResponse(c, errors.New(status.RequireLogin), nil)
		c.Abort()
		return
	}
}

// RequireOwner 拥有者才能访问
func RequireOwner(c *gin.Context) {
	u := session.Get(c)
	if u == nil {
		response.HandleResponse(c, errors.New(status.RequireLogin), nil)
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

// WithScope 使用 scope_token 校验用户登录状态，并替换成对应的用户 token
func WithScope(scope string) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		claims, err := jwt.ValidateUserJwt(c.GetHeader(types.ScopeTokenKey))
		if err != nil {
			c.Next()
			return
		}

		if !strings.Contains(claims.Scope, scope) {
			c.Next()
			return
		}

		if user, err := entity.GetUserByID(claims.UID); err == nil {
			c.Request.Header.Add(types.SATokenKey, user.Token)
			if entity.IsAreaOwner(claims.UID) {
				c.Request.Header.Add(types.RoleKey, types.OwnerRole)
			}

		}
		c.Next()
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
