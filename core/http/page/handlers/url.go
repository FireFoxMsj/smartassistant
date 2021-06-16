package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
)

func RegisterPageRouter(r gin.IRouter) {
	pageGroup := r.Group("/pages", middleware.DefaultMiddleware())
	{
		// TODO:判断当前用户是否有创建场景的权限
		pageGroup.GET("create_scene", CreateSceneGuideRedirect)
	}
}
