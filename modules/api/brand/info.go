package brand

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// Brand 品牌信息
type Brand struct {
	cloud.Brand
	Plugins  []Plugin `json:"plugins"`
	IsAdded  bool     `json:"is_added"`  // 是否已添加
	IsNewest bool     `json:"is_newest"` // 是否是最新
}

type Plugin struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Brand    string `json:"brand"`
	Info     string `json:"info"`
	IsAdded  bool   `json:"is_added"`
	IsNewest bool   `json:"is_newest"`
}

// brandInfoReq 品牌详情接口请求参数
type brandInfoReq struct {
	Name string `uri:"name"`
}

// GetBrandInfo 获取品牌详情
func GetBrandInfo(name string) (brand Brand, err error) {
	brand.Plugins = make([]Plugin, 0)

	brand = Brand{
		Brand: cloud.Brand{
			LogoURL: "", // TODO 本地图片
			Name:    name,
		},
	}

	var installedPlgs []entity.PluginInfo
	installedPlgs, err = entity.GetInstalledPlugins()
	if err != nil {
		return
	}

	installedPlgMap := make(map[string]entity.PluginInfo)
	for _, p := range installedPlgs {
		installedPlgMap[p.PluginID] = p
	}
	brandInfo, err := cloud.GetBrandInfo(name)
	// 请求sc失败则读取本地信息
	if err != nil {
		logrus.Error(err)

		for _, p := range installedPlgs {
			pp := Plugin{
				ID:       p.PluginID,
				Version:  p.Version,
				Brand:    p.Brand,
				IsAdded:  false,
				IsNewest: false,
			}
			brand.IsNewest = true
			brand.IsAdded = true
			brand.Plugins = append(brand.Plugins, pp)
		}
		return
	}
	brand.LogoURL = brandInfo.LogoURL
	for _, p := range brandInfo.Plugins {
		pp := Plugin{
			ID:      p.Domain,
			Name:    p.Name,
			Version: p.Version,
			Brand:   p.Brand,
			Info:    p.Intro,
		}
		_, pp.IsAdded = installedPlgMap[p.Domain]
		if pp.IsAdded {
			pp.IsNewest = p.Version == installedPlgMap[p.Domain].Version
			brand.IsAdded = true
		}
		brand.Plugins = append(brand.Plugins, pp)
	}

	brand.PluginAmount = len(brand.Plugins)
	return
}

// InfoResp 品牌详情接口返回数据
type InfoResp struct {
	Brand Brand `json:"brand"`
}

// Info 用于处理品牌详情接口的请求
func Info(c *gin.Context) {
	var (
		req  brandInfoReq
		resp InfoResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindUri(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	var brand Brand
	if brand, err = GetBrandInfo(req.Name); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	} else {
		resp.Brand = brand
	}
}
