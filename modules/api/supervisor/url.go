package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

func RegisterSupervisorRouter(r gin.IRouter) {
	supervisorGroup := r.Group("supervisor", middleware.RequireOwner)
	{
		supervisorGroup.GET("backups", ListBackup)
		supervisorGroup.POST("backups", AddBackup)
		supervisorGroup.DELETE("backups", DeleteBackup)
		supervisorGroup.POST("backups/restore", Restore)
	}
	r.GET("supervisor/update", middleware.RequireAccount, middleware.RequirePermission(getSwUpgradePermission()), UpdateInfo)
	r.POST("supervisor/update", middleware.RequireAccount, middleware.RequirePermission(getSwUpgradePermission()), Update)
}

// getSwUpgradePermission 获取软件升级权限
func getSwUpgradePermission() (p types.Permission) {
	device, err := entity.GetSaDevice()
	if err != nil {
		logger.Error(err)
		return
	}

	return types.NewDeviceManage(device.ID, "软件升级", types.SoftwareUpgrade)
}
