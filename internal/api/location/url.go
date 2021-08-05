// Package location 房间
package location

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/device"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
	"github.com/zhiting-tech/smartassistant/internal/types"
)

// RegisterLocationRouter 注册与房间相关的路由及其处理函数
func RegisterLocationRouter(r gin.IRouter) {
	locationGroup := r.Group("locations", middleware.RequireAccount)

	locationGroup.PUT(":id", middleware.RequirePermission(types.LocationUpdateName), UpdateLocation)
	locationGroup.DELETE(":id", middleware.RequirePermission(types.LocationDel), DelLocation)
	locationGroup.GET(":id", middleware.RequirePermission(types.LocationGet), InfoLocation)
	locationGroup.GET(":id/devices", device.ListLocationDevices)

	locationGroup.GET("", ListLocation)
	locationGroup.POST("", middleware.RequirePermission(types.LocationAdd), AddLocation)
	locationGroup.PUT("", middleware.RequirePermission(types.LocationUpdateOrder), LocationOrder)

	r.GET("location_tmpl", ListDefaultLocation)
}
