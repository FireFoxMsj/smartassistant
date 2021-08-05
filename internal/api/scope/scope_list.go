package scope

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
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
	for k, v := range scopes {
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
