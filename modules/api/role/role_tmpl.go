package role

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
)

// getAllPermissions 用于处理权限模板接口的请求
func getAllPermissions(c *gin.Context) {

	var (
		resp roleGetResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	ps, err := getPermissionsWithDevices(c)
	if err != nil {
		return
	}
	resp.Role.Permissions = &ps
}
