package brand

import (
	errors2 "errors"
	"net/http"

	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/plugin"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// Branch 品牌信息
type Brand struct {
	LogoURL        string           `json:"logo_url"`
	Name           string           `json:"name"`
	Plugins        []Plugin         `json:"plugins"`
	IsAdded        bool             `json:"is_added"`  // 是否已添加
	IsNewest       bool             `json:"is_newest"` // 是否是最新
	SupportDevices []*plugin.Device `json:"support_devices"`
}

type Plugin struct {
	plugin.Plugin
	IsAdded  bool `json:"is_added"`
	IsNewest bool `json:"is_newest"`
}

// brandInfoReq 品牌详情接口请求参数
type brandInfoReq struct {
	Name string `uri:"name"`
}

func getBrandFromPlgs(req *http.Request, brandName string, plgs []*plugin.Plugin) (brand Brand, err error) {
	brand = Brand{
		Name: brandName,
	}
	m := plugin.GetManager()
	err = errors2.New("brand not exist")
	for _, plg := range plgs {
		if plg.Brand == brandName {
			err = nil
			isAdded, isNewest := m.PluginStatus(plg.ID)
			brand.Plugins = append(brand.Plugins, Plugin{*plg, isAdded, isNewest})
			if isAdded {
				brand.IsAdded = true
			}
			if isNewest {
				brand.IsNewest = true
			}
			brand.LogoURL = plg.LogoURLWithRequest(req)
			brand.SupportDevices = append(brand.SupportDevices, plg.SupportDevices...)
		}
	}
	return
}

// GetBrandInfo 获取品牌详情
func GetBrandInfo(req *http.Request, name string) (brand Brand, err error) {
	plgs := plugin.GetManager().ListPlugin()
	return getBrandFromPlgs(req, name, plgs)
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

	if brand, err := GetBrandInfo(c.Request, req.Name); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	} else {
		resp.Brand = brand
	}
}
