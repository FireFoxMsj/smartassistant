package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
)

func RegisterRoleRouter(r gin.IRouter) {
	r.GET("role_tmpl", middleware.RequireAccount, getAllPermissions) // 所有权限的模板
	roleGroup := r.Group("roles", middleware.RequireAccount)
	roleGroup.GET("", middleware.RequirePermission(permission.RoleGet), roleList)
	roleGroup.POST("", middleware.RequirePermission(permission.RoleAdd), roleUpdate)
	roleGroup.GET(":id", roleGet)
	roleGroup.PUT(":id", middleware.RequirePermission(permission.RoleUpdate), roleUpdate)
	roleGroup.DELETE(":id", middleware.RequirePermission(permission.RoleDel), roleDel)
}
