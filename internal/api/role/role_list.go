package role

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
)

// roleListResp 角色列表接口返回数据
type roleListResp struct {
	Roles []roleInfo `json:"roles"`
}

// roleList 用于处理角色列表接口的请求
func roleList(c *gin.Context) {
	var (
		resp roleListResp
		err  error
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	resp.Roles = make([]roleInfo, 0)
	roles, err := entity.GetRoles()
	if err != nil {
		return
	}

	// 返回拥有者的角色信息
	resp.Roles = append(resp.Roles, roleInfo{
		RoleInfo: entity.RoleInfo{
			ID:   entity.OwnerRoleID,
			Name: entity.Owner,
		},
		IsManager: true,
	})

	for _, r := range roles {
		roleConf := roleInfo{
			RoleInfo: entity.RoleInfo{
				ID:   r.ID,
				Name: r.Name,
			},
			IsManager: r.IsManager,
		}
		resp.Roles = append(resp.Roles, roleConf)
	}

	return
}
