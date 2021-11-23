package auth

import "github.com/gin-gonic/gin"

func InitAuthRouter(r gin.IRouter) {
	aGroup := r.Group("oauth")
	aGroup.POST("access_token", GetToken)
	aGroup.GET("authorize_code", GetAuthorizeCode) // 获取授权码
}
