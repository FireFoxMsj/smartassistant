// Package role 用户角色
package role

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
	"github.com/zhiting-tech/smartassistant/internal/types"
)

// RegisterRoleRouter 注册与角色相关的路由及其处理函数
func RegisterRoleRouter(r gin.IRouter) {
	r.GET("role_tmpl", middleware.RequireAccount, getAllPermissions) // 所有权限的模板
	roleGroup := r.Group("roles", middleware.RequireAccount)
	roleGroup.GET("", middleware.RequirePermission(types.RoleGet), roleList)
	roleGroup.POST("", middleware.RequirePermission(types.RoleAdd), roleUpdate)
	roleGroup.GET(":id", roleGet)
	roleGroup.PUT(":id", middleware.RequirePermission(types.RoleUpdate), roleUpdate)
	roleGroup.DELETE(":id", middleware.RequirePermission(types.RoleDel), roleDel)
}
