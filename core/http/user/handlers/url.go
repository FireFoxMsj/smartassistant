package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/role/handlers"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
)

func RegisterUserRouter(r gin.IRouter) {
	userGroup := r.Group("/users", middleware.RequireAccount)
	{
		userGroup.GET(":id", InfoUser)
		userGroup.PUT(":id", UpdateUser)
		userGroup.GET("", ListUser)
		userGroup.DELETE(":id", middleware.RequirePermission(permission.AreaDelMember), DelUser)
		userGroup.POST("/:id/invitation/code", middleware.RequirePermission(permission.AreaGetCode), GetInvitationCode)
		userGroup.GET(":id/permissions", handlers.UserPermissions)
	}
	invitationGroup := r.Group("/invitation")
	{
		invitationGroup.POST("/check", CheckQrCode)
	}
}
