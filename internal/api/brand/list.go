package brand

import (
	"net/http"

	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/plugin"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// 0,获取所有品牌;1,获取已添加的品牌列表
const (
	typeTotal = 0
	typeAdded = 1
)

// BrandInfo 品牌信息
type BrandInfo struct {
	LogoURL      string   `json:"logo_url"`
	Name         string   `json:"name"`
	PluginAmount int      `json:"plugin_amount"` // 插件数量
	Plugins      []Plugin `json:"plugins"`
	IsAdded      bool     `json:"is_added"`  // 是否已添加
	IsNewest     bool     `json:"is_newest"` // 是否是最新
}

// Resp 品牌列表接口返回数据
type Resp struct {
	Brands []BrandInfo `json:"brands"`
}

// listBrandsReq 品牌列表接口请求参数
type listBrandsReq struct {
	Type int `form:"type"` // 0全部1已安装
}

// pluginsToBrands 将插件列表按品牌划分
func pluginsToBrands(req *http.Request, plugins []*plugin.Plugin) map[string]BrandInfo {
	m := plugin.GetManager()
	brandMap := make(map[string]BrandInfo)
	for _, plg := range plugins {
		brand, ok := brandMap[plg.Brand]
		if !ok {
			brand = BrandInfo{
				LogoURL: plg.LogoURLWithRequest(req),
				Name:    plg.Brand,
			}
		}
		brand.PluginAmount += 1

		isAdded, isNewest := m.PluginStatus(plg.ID)
		if isAdded {
			brand.IsAdded = true
		}
		if isNewest {
			brand.IsNewest = true
		}
		brand.Plugins = append(brand.Plugins, Plugin{*plg, isAdded, isNewest})
		brandMap[plg.Brand] = brand
	}
	return brandMap
}

func ListBrands(req *http.Request, isAdded bool) (brands []BrandInfo) {
	plgs := plugin.GetManager().ListPlugin()
	if len(plgs) == 0 {
		return
	}
	brandMap := pluginsToBrands(req, plgs)
	for _, b := range brandMap {
		if isAdded && !b.IsAdded {
			continue
		}
		brands = append(brands, b)
	}
	return
}

// List 用于处理品牌列表接口的请求
func List(c *gin.Context) {
	var (
		resp Resp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()
	var req listBrandsReq
	if err = c.ShouldBind(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	resp = Resp{
		Brands: ListBrands(c.Request, req.Type == typeAdded),
	}
}
