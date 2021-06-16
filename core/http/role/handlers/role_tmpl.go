package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

func getAllPermissions(c *gin.Context) {

	var (
		resp roleGetResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	ps, err := getPermissions()
	if err != nil {
		return
	}
	resp.Role.Permissions = &ps
}
