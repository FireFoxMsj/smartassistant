// Package session 用户登录登出
package session

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
)

// InitSessionRouter 注册与用户登录相关的路由及其处理函数
func InitSessionRouter(router gin.IRouter) {
	sessionGroup := router.Group("/sessions", middleware.DefaultMiddleware())
	{
		sessionGroup.POST("/login", Login)
		sessionGroup.POST("/logout", Logout)
	}
}
