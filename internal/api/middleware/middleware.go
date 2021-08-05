// Package middleware GIN 框架中间件
package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/cache"
	"github.com/zhiting-tech/smartassistant/internal/utils/jwt"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strings"
)

// RequireAccount 用户需要登录才可访问对应的接口
func RequireAccount(c *gin.Context) {

	// 校验请求头verification-code是否有值
	code := c.GetHeader(types.VerificationKey)
	val := cache.GetValWithCode(code)
	if code != "" && val != "" {
		c.Request.Header.Set(types.SATokenKey, val)
		c.Next()
		// 验证成功后删除code
		cache.GetCache().Delete(code)
		return
	}

	u := session.Get(c)
	if u == nil {
		response.HandleResponse(c, errors.New(status.RequireLogin), nil)
		c.Abort()
		return
	}
}

// WithScope 使用 scope_token 校验用户登录状态，并替换成对应的用户 token
func WithScope(scope string) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		claims, err := jwt.ValidateUserJwt(c.GetHeader(types.ScopeTokenKey))
		if err != nil {
			c.Next()
			return
		}
		scopes, ok := claims["scopes"].(string)
		if !(ok && strings.Contains(scopes, scope)) {
			c.Next()
			return
		}

		if uid, ok := claims["uid"].(float64); ok {
			if user, err := entity.GetUserByID(int(uid)); err == nil {
				c.Request.Header.Add(types.SATokenKey, user.Token)
				if entity.IsSAOwner(int(uid)) {
					c.Request.Header.Add(types.RoleKey, types.OwnerRole)
				}

			}
		}
		c.Next()
	}
}

// RequireToken 使用token验证身份，不依赖cookies.
func RequireToken(c *gin.Context) {

	queryToken := c.Query("token")
	if queryToken != "" {
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
