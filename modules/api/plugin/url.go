package plugin

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

// RegisterPluginRouter 插件
func RegisterPluginRouter(r gin.IRouter) {
	pluginGroup := r.Group("plugins")
	pluginAuthGroup := pluginGroup.Use(middleware.RequireAccount)

	pluginGroup.GET(":id", PluginInfo)
	pluginAuthGroup.POST("", UploadPlugin)
}
