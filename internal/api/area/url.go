// Package area 公司/家庭
package area

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
	"github.com/zhiting-tech/smartassistant/internal/types"
)

// RegisterAreaRouter 用于注册与家庭相关的路由及其处理函数
func RegisterAreaRouter(r gin.IRouter) {
	areaAuthGroup := r.Group("areas", middleware.WithScope("area"), middleware.RequireAccount)
	areaAuthGroup.GET("", ListArea)
	areaAuthGroup.PUT(":id", middleware.RequirePermission(types.AreaUpdateName), UpdateArea)
	areaAuthGroup.DELETE(":id", DelArea)
	areaAuthGroup.GET(":id", InfoArea)
	areaAuthGroup.DELETE(":id/users/:user_id", middleware.RequireAccount, QuitArea)

	r.POST("/sync", middleware.RequireAccount, DataSync)
}
