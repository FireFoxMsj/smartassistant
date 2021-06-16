package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

const (
	typeTotal = 0
	typeAdded = 1
)

type BrandInfo struct {
	LogoURL      string          `json:"logo_url"`
	Name         string          `json:"name"`
	PluginAmount int             `json:"plugin_amount"` // 插件数量
	Plugins      []plugin.Plugin `json:"plugins"`
	IsAdded      bool            `json:"is_added"`  // 是否已添加
	IsNewest     bool            `json:"is_newest"` // 是否是最新
}

type Resp struct {
	Brands []BrandInfo `json:"brands"`
}

type listBrandsReq struct {
	Type int `form:"type"` // 0全部1已安装
}

// pluginsToBrands 将插件列表按品牌划分
func pluginsToBrands(plugins []plugin.Plugin) map[string]BrandInfo {

	brandMap := make(map[string]BrandInfo)
	for _, plg := range plugins {
		brand, ok := brandMap[plg.Brand]
		if !ok {
			brand = BrandInfo{
				LogoURL: "www.baidu.com/brand_logo.jpg",
				Name:    plg.Brand,
			}
		}
		brand.PluginAmount += 1
		brand.Plugins = append(brand.Plugins, plg)
		if plg.IsAdded {
			brand.IsAdded = true
		}
		if plg.IsNewest {
			brand.IsNewest = true
		}
		brandMap[plg.Brand] = brand
	}
	return brandMap
}

func ListBrands(isAdded bool) (brands []BrandInfo) {
	plgs := plugin.GetPlugins()

	brandMap := pluginsToBrands(plgs)
	for _, b := range brandMap {
		if isAdded && !b.IsAdded {
			continue
		}
		brands = append(brands, b)
	}
	return
}

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
		Brands: ListBrands(req.Type == typeAdded),
	}
}
