package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

func RequireAccount(c *gin.Context) {
	u := session.Get(c)
	if u == nil {
		response.HandleResponse(c, errors.New(errors.RequireLogin), nil)
		c.Abort()
		return
	}
}

// RequireToken 使用token验证身份，不依赖cookies.
func RequireToken(c *gin.Context) {

	queryToken := c.Query("token")
	if queryToken != "" {
		c.Request.Header.Add(core.SATokenKey, queryToken)
	}
	if session.GetUserByToken(c) != nil {
		c.Next()
		return
	}
	response.HandleResponse(c, errors.New(errors.RequireLogin), nil)
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
func RequirePermission(p permission.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := session.Get(c)
		if u != nil {
			if orm.JudgePermit(u.UserID, p) {
				c.Next()
				return
			}
		}
		err := errors.New(errors.Deny)
		response.HandleResponse(c, err, nil)
		c.Abort()

	}
}
