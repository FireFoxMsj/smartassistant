package api

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/area"
	"github.com/zhiting-tech/smartassistant/internal/api/brand"
	"github.com/zhiting-tech/smartassistant/internal/api/cloud"
	"github.com/zhiting-tech/smartassistant/internal/api/device"
	"github.com/zhiting-tech/smartassistant/internal/api/location"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
	"github.com/zhiting-tech/smartassistant/internal/api/page"
	"github.com/zhiting-tech/smartassistant/internal/api/role"
	"github.com/zhiting-tech/smartassistant/internal/api/scene"
	"github.com/zhiting-tech/smartassistant/internal/api/scope"
	"github.com/zhiting-tech/smartassistant/internal/api/session"
	"github.com/zhiting-tech/smartassistant/internal/api/user"
	"github.com/zhiting-tech/smartassistant/pkg/reverseproxy"
)

// loadModules 注册路由及其处理函数
func loadModules(r gin.IRouter) {
	r.Use(middleware.DefaultMiddleware())
	location.RegisterLocationRouter(r)
	brand.RegisterBrandRouter(r)
	device.RegisterDeviceRouter(r)
	area.RegisterAreaRouter(r)
	user.RegisterUserRouter(r)
	scope.RegisterScopeRouter(r)
	role.RegisterRoleRouter(r)
	session.InitSessionRouter(r)
	scene.InitSceneRouter(r)
	page.RegisterPageRouter(r)
	cloud.InitCloudRouter(r)
	r.Any("/plugin/:plugin/*path", middleware.WithScope("user"), reverseproxy.ProxyToPlugin)
}
