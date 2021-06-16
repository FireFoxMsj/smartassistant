package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
)

func InitSceneRouter(r gin.IRouter) {
	sceneGroup := r.Group("scenes", middleware.RequireAccount)
	{
		sceneGroup.POST("", CreateScene)
		sceneGroup.DELETE(":id", DeleteScene)
		sceneGroup.PUT(":id", middleware.RequirePermission(permission.SceneUpdate), UpdateScene)
		sceneGroup.GET("", ListScene)
		sceneGroup.GET(":id", InfoScene)
		sceneGroup.POST(":id/execute", ExecuteScene)
	}

	r.GET("scene_logs", ListSceneTaskLog)
}
