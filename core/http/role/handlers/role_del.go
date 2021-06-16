package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

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

	if err = orm.DeleteRole(roleId); err != nil {
		return
	}
}
