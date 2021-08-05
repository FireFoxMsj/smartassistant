// Package scope 用户 Scope Token
package scope

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
)

// RegisterScopeRouter scope token 路由注册
func RegisterScopeRouter(r gin.IRouter) {
	scopeGroup := r.Group("scopes", middleware.RequireAccount)
	scopeGroup.GET("", scopeList)
	scopeGroup.POST("token", scopeToken)
}
