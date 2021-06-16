package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
)

func RegisterDeviceRouter(r gin.IRouter) {
	deviceGroup := r.Group("devices")
	deviceGroup.POST("", AddDevice)

	deviceAuthGroup := r.Group("devices", middleware.RequireAccount)
	deviceAuthGroup.GET("", ListAllDevice)
	deviceAuthGroup.PUT(":id", UpdateDevice)
	deviceAuthGroup.GET(":id", InfoDevice)
	deviceAuthGroup.DELETE(":id", DelDevice)
	r.GET("/check", CheckSaDevice)
}
