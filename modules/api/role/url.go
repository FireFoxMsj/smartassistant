// Package role 用户角色
package role

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
)

// RegisterRoleRouter 注册与角色相关的路由及其处理函数
func RegisterRoleRouter(r gin.IRouter) {
	r.GET("role_tmpl", middleware.RequireAccount, getAllPermissions) // 所有权限的模板
	roleGroup := r.Group("roles", middleware.RequireAccount)
	roleGroup.GET("", middleware.RequirePermission(types.RoleGet), roleList)
	roleGroup.POST("", middleware.RequirePermission(types.RoleAdd), roleUpdate)
	roleGroup.GET(":id", requireBelongsToUser, roleGet)
	roleGroup.PUT(":id", requireBelongsToUser, middleware.RequirePermission(types.RoleUpdate), roleUpdate)
	roleGroup.DELETE(":id", requireBelongsToUser, middleware.RequirePermission(types.RoleDel), roleDel)
}

// requireBelongsToUser 操作角色需要角色属于用户的家庭
func requireBelongsToUser(c *gin.Context) {
	user, err := entity.GetUserByID(session.Get(c).UserID)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.InternalServerErr), nil)
		c.Abort()
		return
	}

	roleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.BadRequest), nil)
		c.Abort()
		return
	}

	role, err := entity.GetRoleByID(roleID)
	if err != nil {
		response.HandleResponse(c, errors.Wrap(err, errors.InternalServerErr), nil)
		c.Abort()
		return
	}

	if role.AreaID == user.AreaID {
		c.Next()
	} else {
		response.HandleResponse(c, errors.New(status.Deny), nil)
		c.Abort()
	}
}
