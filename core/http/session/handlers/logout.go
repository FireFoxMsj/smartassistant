package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

func Logout(c *gin.Context) {
	defer func() {
		response.HandleResponse(c, nil, nil)
	}()

	session.Logout(c)
}
