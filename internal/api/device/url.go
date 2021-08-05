// Package device 设备，包括SA状态
package device

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
)

// RegisterDeviceRouter 注册与设备相关的路由及其处理函数
func RegisterDeviceRouter(r gin.IRouter) {
	deviceGroup := r.Group("devices")
	deviceGroup.POST("", AddDevice)

	deviceAuthGroup := r.Group("devices", middleware.RequireAccount)
	deviceAuthGroup.GET("", ListAllDevice)
	deviceAuthGroup.PUT(":id", UpdateDevice)
	deviceAuthGroup.GET(":id", InfoDevice)
	deviceAuthGroup.DELETE(":id", DelDevice)

	// 检查SA是否已绑定
	r.GET("/check", CheckSaDevice)
	// 检查SA是否对应的云端同步下来的家庭
	r.POST("/check", IsAccessAllow)

}
