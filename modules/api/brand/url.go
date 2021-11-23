// Package brand 品牌
package brand

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

// RegisterBrandRouter 注册与品牌相关的路由及其处理函数
func RegisterBrandRouter(r gin.IRouter) {
	brandGroup := r.Group("brands", middleware.RequireAccount)
	brandGroup.GET("", List)
	brandGroup.GET(":name", Info)

	// 插件的安装、更新、删除
	brandGroup.POST(":brand_name/plugins", UpdatePlugin)
	brandGroup.DELETE(":brand_name/plugins", DelPlugins)
}
