// Package cloud 智汀云对接
package cloud

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

func InitCloudRouter(r gin.IRouter) {
	r.POST("cloud/bind", middleware.RequireAccount, bindCloud)
	r.POST("cloud/migration", middleware.RequireOwner, AreaMigration)
}
