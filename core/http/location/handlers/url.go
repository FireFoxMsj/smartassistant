package handlers

import (
	"github.com/gin-gonic/gin"

	handlers3 "gitlab.yctc.tech/root/smartassistent.git/core/http/device/handlers"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
)

func RegisterLocationRouter(r gin.IRouter) {
	locationGroup := r.Group("locations", middleware.RequireAccount)

	locationGroup.PUT(":id", middleware.RequirePermission(permission.LocationUpdateName), UpdateLocation)
	locationGroup.DELETE(":id", middleware.RequirePermission(permission.LocationDel), DelLocation)
	locationGroup.GET(":id", middleware.RequirePermission(permission.LocationGet), InfoLocation)
	locationGroup.GET(":id/devices", handlers3.ListLocationDevices)

	locationGroup.GET("", ListLocation)
	locationGroup.POST("", middleware.RequirePermission(permission.LocationAdd), AddLocation)
	locationGroup.PUT("", middleware.RequirePermission(permission.LocationUpdateOrder), LocationOrder)

	r.GET("location_tmpl", ListDefaultLocation)
}
