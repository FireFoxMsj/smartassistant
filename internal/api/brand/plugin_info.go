package brand

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/plugin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// PluginInfoResp 插件详情接口返回数据
type PluginInfoResp struct {
	Plugin Plugin `json:"plugin"`
}

// PluginInfoReq 插件详情接口请求参数
type PluginInfoReq struct {
	PluginID string `uri:"id"`
}

// PluginInfo 用于处理插件详情接口的请求
func PluginInfo(c *gin.Context) {
	var (
		err  error
		req  PluginInfoReq
		resp PluginInfoResp
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindUri(&req); err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	plg, err := plugin.GetManager().GetPlugin(req.PluginID)
	if err != nil || plg == nil {
		return
	}
	isAdded, isNewest := plugin.GetManager().PluginStatus(plg.ID)
	resp.Plugin = Plugin{*plg, isAdded, isNewest}
}
