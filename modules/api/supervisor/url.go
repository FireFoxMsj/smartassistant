package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

func RegisterSupervisorRouter(r gin.IRouter) {
	supervisorGroup := r.Group("supervisor", middleware.RequireAccount)
	{
		supervisorGroup.POST("backup", AddBackup)
		supervisorGroup.POST("restore", Restore)
	}
}
