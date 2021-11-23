package scope

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/types"
)

type scopeItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type scopeListResp struct {
	Scopes []scopeItem `json:"scopes"`
}

// scopeList 返回预定义的范围权限列表
func scopeList(c *gin.Context) {
	scp := make([]scopeItem, 0)
	for k, v := range types.Scopes {
		if k == "device" {
			continue
		}
		scp = append(scp, scopeItem{
			Name:        k,
			Description: v,
		})
	}
	resp := scopeListResp{
		Scopes: scp,
	}
	response.HandleResponse(c, nil, resp)
}
