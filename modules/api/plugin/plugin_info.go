package plugin

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/brand"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// PluginInfoResp 插件详情接口返回数据
type PluginInfoResp struct {
	Plugin brand.Plugin `json:"plugin"`
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

	plg, err := plugin.GetGlobalManager().Get(req.PluginID)
	if err != nil || plg == nil {
		return
	}
	resp.Plugin = brand.Plugin{Plugin: *plg, IsAdded: plg.IsAdded(), IsNewest: plg.IsNewest()}
	resp.Plugin.DownloadURL = plugin.ArchiveURL(plg.ID, c.Request)
}
