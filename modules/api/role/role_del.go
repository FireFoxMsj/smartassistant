package role

import (
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"

	"github.com/gin-gonic/gin"
)

// roleDel 用于处理删除角色接口的请求
func roleDel(c *gin.Context) {
	var (
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	v := c.Param("id")
	roleId, err := strconv.Atoi(v)
	if err != nil {
		return
	}

	if err = entity.DeleteRole(roleId); err != nil {
		return
	}
}
