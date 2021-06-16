package handlers

import (
	errors2 "errors"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type Brand struct {
	LogoURL        string          `json:"logo_url"`
	Name           string          `json:"name"`
	Plugins        []plugin.Plugin `json:"plugins"`
	IsAdded        bool            `json:"is_added"`  // 是否已添加
	IsNewest       bool            `json:"is_newest"` // 是否是最新
	SupportDevices []plugin.Device `json:"support_devices"`
}

type brandInfoReq struct {
	Name string `uri:"name"`
}

func getBrandFromPlgs(brandName string, plgs []plugin.Plugin) (brand Brand, err error) {
	brand = Brand{
		LogoURL: "www.baidu.com/brand_logo.jpg",
		Name:    brandName,
	}
	err = errors2.New("brand not exist")
	for _, plg := range plgs {
		if plg.Brand == brandName {
			err = nil
			brand.Plugins = append(brand.Plugins, plg)
			if plg.IsAdded {
				brand.IsAdded = true
			}
			if plg.IsNewest {
				brand.IsNewest = true
			}
			brand.SupportDevices = append(brand.SupportDevices, plg.SupportDevices...)
		}
	}
	return
}

// GetBrandInfo 获取品牌详情
func GetBrandInfo(name string) (brand Brand, err error) {
	plgs := plugin.GetPlugins()
	return getBrandFromPlgs(name, plgs)
}

type InfoResp struct {
	Brand Brand `json:"brand"`
}

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

	if brand, err := GetBrandInfo(req.Name); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	} else {
		resp.Brand = brand
	}
}
