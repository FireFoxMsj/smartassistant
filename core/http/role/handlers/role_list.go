package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type roleListResp struct {
	Roles []roleInfo `json:"roles"`
}

func roleList(c *gin.Context) {
	var (
		resp roleListResp
		err  error
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	resp.Roles = make([]roleInfo, 0)
	roles, err := orm.GetRoles()
	if err != nil {
		return
	}

	for _, r := range roles {
		roleConf := roleInfo{
			RoleInfo: orm.RoleInfo{
				ID:   r.ID,
				Name: r.Name,
			},
			IsManager: r.IsManager,
		}
		resp.Roles = append(resp.Roles, roleConf)
	}

	return
}
