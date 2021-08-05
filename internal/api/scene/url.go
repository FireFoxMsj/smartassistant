// Package scene 设备场景
package scene

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
	"github.com/zhiting-tech/smartassistant/internal/types"
)

// InitSceneRouter 注册与场景相关的路由及其处理函数
func InitSceneRouter(r gin.IRouter) {
	sceneGroup := r.Group("scenes", middleware.RequireAccount)
	{
		sceneGroup.POST("", CreateScene)
		sceneGroup.DELETE(":id", DeleteScene)
		sceneGroup.PUT(":id", middleware.RequirePermission(types.SceneUpdate), UpdateScene)
		sceneGroup.GET("", ListScene)
		sceneGroup.GET(":id", InfoScene)
		sceneGroup.POST(":id/execute", ExecuteScene)
	}

	r.GET("scene_logs", ListSceneTaskLog)
}
