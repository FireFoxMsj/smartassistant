package session

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
)

// Logout 用于处理用户登出接口的请求
func Logout(c *gin.Context) {
	defer func() {
		response.HandleResponse(c, nil, nil)
	}()

	session.Logout(c)
}
