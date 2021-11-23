package cloud

import (
	"fmt"
	"io/ioutil"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/config"
)

type DeviceType string
type DeviceSubType string

const (
	TypeLight          DeviceType = "light"           // 灯
	TypeSwitch         DeviceType = "switch"          // 开关
	TypeOutlet         DeviceType = "outlet"          // 插座
	TypeRoutingGateway DeviceType = "routing_gateway" // 路由网关
	TypeSecurity       DeviceType = "security"        // 安防
)

type Brand struct {
	LogoURL      string `json:"logo_url"`
	Name         string `json:"name"`
	PluginAmount int    `json:"plugin_amount"` // 插件数量
}

type Plugin struct {
	ID      int    `json:"id"`
	Domain  string `json:"domain"`
	Name    string `json:"name"`
	Image   string `json:"image"`
	Version string `json:"version"`
	Brand   string `json:"brand"`
	Intro   string `json:"intro"`
}

type Device struct {
	Model string     `json:"model" `
	Name  string     `json:"name"`
	Type  DeviceType `json:"type"` // 设备类型

	Logo         string `json:"logo" `        // 设备logo相对路径
	Control      string `json:"control"`      // 设备控制页面相对路径
	Provisioning string `json:"provisioning"` // 设备置网页面相对路径
}

// SaveBrandLogos TODO 保存所有品牌logo
func SaveBrandLogos() {

}

func GetBrands() (brands []Brand, err error) {

	url := fmt.Sprintf("%s/common/brands", config.GetConf().SmartCloud.URL())
	logrus.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	any := jsoniter.Get(data, "data", "list")
	any.ToVal(&brands)
	err = any.LastError()
	return
}

func GetBrandsMap() (brandsMap map[string]Brand, err error) {
	brands, err := GetBrands()
	if err != nil {
		return
	}
	brandsMap = make(map[string]Brand)
	for _, brand := range brands {
		brandsMap[brand.Name] = brand
	}
	return
}

type BrandInfo struct {
	Brand
	Plugins []Plugin `json:"plugins"`
}

func GetBrandInfo(brandName string) (brand BrandInfo, err error) {
	url := fmt.Sprintf("%s/common/brands/name/%s", config.GetConf().SmartCloud.URL(), brandName)
	logrus.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	any := jsoniter.Get(data, "data")
	logrus.Println(any.ToString())
	any.ToVal(&brand)
	err = any.LastError()
	return
}

func GetPlugins() (plugins []Plugin, err error) {
	url := fmt.Sprintf("%s/common/plugins", config.GetConf().SmartCloud.URL())
	logrus.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	any := jsoniter.Get(data, "data", "list")
	any.ToVal(&plugins)
	err = any.LastError()
	return
}

func GetPlugin(domain string) (plugin Plugin, err error) {
	url := fmt.Sprintf("%s/common/plugins/domain/%s",
		config.GetConf().SmartCloud.URL(), domain)
	logrus.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	any := jsoniter.Get(data, "data", "plugin")
	any.ToVal(&plugin)
	err = any.LastError()
	return
}
