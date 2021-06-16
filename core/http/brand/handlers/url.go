package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
)

func RegisterBrandRouter(r gin.IRouter) {
	brandGroup := r.Group("brands", middleware.RequireAccount)
	brandGroup.GET("", List)
	brandGroup.GET(":name", Info)
}
