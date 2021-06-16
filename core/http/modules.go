package http

import (
	"github.com/gin-gonic/gin"
	scenes "gitlab.yctc.tech/root/smartassistent.git/core/http/scene/handlers"

	area "gitlab.yctc.tech/root/smartassistent.git/core/http/area/handlers"
	brand "gitlab.yctc.tech/root/smartassistent.git/core/http/brand/handlers"
	device "gitlab.yctc.tech/root/smartassistent.git/core/http/device/handlers"
	location "gitlab.yctc.tech/root/smartassistent.git/core/http/location/handlers"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
	page "gitlab.yctc.tech/root/smartassistent.git/core/http/page/handlers"
	role "gitlab.yctc.tech/root/smartassistent.git/core/http/role/handlers"
	sessions "gitlab.yctc.tech/root/smartassistent.git/core/http/session/handlers"
	user "gitlab.yctc.tech/root/smartassistent.git/core/http/user/handlers"
)

func LoadModules(r gin.IRouter) {
	r.Use(middleware.DefaultMiddleware())
	location.RegisterLocationRouter(r)
	brand.RegisterBrandRouter(r)
	device.RegisterDeviceRouter(r)
	area.RegisterAreaRouter(r)
	user.RegisterUserRouter(r)
	role.RegisterRoleRouter(r)
	sessions.InitSessionRouter(r)
	scenes.InitSceneRouter(r)
	page.RegisterPageRouter(r)
}
