// Package user 用户管理，权限，邀请别人假如
package user

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
	"github.com/zhiting-tech/smartassistant/internal/api/role"
	"github.com/zhiting-tech/smartassistant/internal/types"
)

// RegisterUserRouter 注册与用户相关的路由及其处理函数
func RegisterUserRouter(r gin.IRouter) {
	userGroup := r.Group("/users", middleware.WithScope("user"), middleware.RequireAccount)
	{
		userGroup.GET(":id", InfoUser)
		userGroup.PUT(":id", UpdateUser)
		userGroup.GET("", ListUser)
		userGroup.DELETE(":id", middleware.RequirePermission(types.AreaDelMember), DelUser)
		userGroup.POST("/:id/invitation/code", middleware.RequirePermission(types.AreaGetCode), GetInvitationCode)
		userGroup.GET(":id/permissions", role.UserPermissions)
		userGroup.PUT("/:id/owner", TransferOwner)
	}
	invitationGroup := r.Group("/invitation")
	{
		invitationGroup.POST("/check", CheckQrCode)
	}
	// 验证码接口，用于第三方云绑定的校验
	r.POST("verification/code", middleware.RequireAccount, GetVerificationCode)
}
