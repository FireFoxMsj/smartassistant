package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
)

func InitSessionRouter(router gin.IRouter) {
	sessionGroup := router.Group("/sessions", middleware.DefaultMiddleware())
	{
		sessionGroup.POST("/login", Login)
		sessionGroup.POST("/logout", Logout)
	}
}
