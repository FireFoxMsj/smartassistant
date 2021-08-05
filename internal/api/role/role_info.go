package role

import (
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"strconv"

	"github.com/gin-gonic/gin"
)

// roleGetResp 角色详情接口请求参数
type roleGetResp struct {
	Role roleInfo `json:"role"`
}

// roleGet 用于处理角色详情接口的请求
func roleGet(c *gin.Context) {
	var (
		resp roleGetResp
		err  error
		role entity.Role
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	v := c.Param("id")
	roleID, err := strconv.Atoi(v)
	if err != nil {
		return
	}

	if role, err = entity.GetRoleByID(roleID); err != nil {
		return
	}

	resp.Role, err = wrapRole(role)
}
