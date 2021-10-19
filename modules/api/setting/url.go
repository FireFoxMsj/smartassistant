package setting

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/middleware"
)

func RegisterSettingRouter(r gin.IRouter) {
	settingGroup := r.Group("setting", middleware.RequireAccount)
	{
		settingGroup.GET("", GetSetting)
		settingGroup.PUT("", middleware.RequireOwner, UpdateSetting)
	}
}
