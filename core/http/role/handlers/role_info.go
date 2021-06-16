package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type roleGetResp struct {
	Role roleInfo `json:"role"`
}

func roleGet(c *gin.Context) {
	var (
		resp roleGetResp
		err  error
		role orm.Role
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	v := c.Param("id")
	roleID, err := strconv.Atoi(v)
	if err != nil {
		return
	}

	if role, err = orm.GetRoleByID(roleID); err != nil {
		return
	}

	resp.Role, err = wrapRole(role)
}
