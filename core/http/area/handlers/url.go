package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
)

func RegisterAreaRouter(r gin.IRouter) {
	areaAuthGroup := r.Group("areas", middleware.RequireAccount)
	areaAuthGroup.GET("", ListArea)
	areaAuthGroup.PUT(":id", middleware.RequirePermission(permission.AreaUpdateName), EditAreaName)
	areaAuthGroup.DELETE(":id", DelArea)
	areaAuthGroup.GET(":id", InfoArea)
	areaAuthGroup.DELETE(":id/users/:user_id", middleware.RequireAccount, QuitArea)

	r.POST("/sync", middleware.RequireAccount, DataSync)
}
