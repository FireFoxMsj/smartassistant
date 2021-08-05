package plugin

import (
	"errors"
	"net/http"

	"github.com/zhiting-tech/smartassistant/internal/plugin/docker"
	"github.com/zhiting-tech/smartassistant/internal/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

var NotExistErr = errors.New("plugin not exist")

type Type string

type Device struct {
	LogoURL string   `json:"logo_url" yaml:"logo_url"`
	Model   string   `json:"model"`
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

type Action struct {
	Cmd           string `yaml:"cmd"`
	Name          string `yaml:"name"`
	Attribute     string `yaml:"attribute"`
	AttributeName string `yaml:"attribute_name"`
	Action        string `yaml:"action"`
}

type Plugin struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Image          docker.Image `json:"image"`
	LogoURL        string       `json:"logo_url"`
	Version        string       `json:"version"`
	Brand          string       `json:"brand"`
	Info           string       `json:"info"`
	DownloadURL    string       `json:"download_url"`
	VisitURL       string       `json:"visit_url"` // TODO 删除
	SupportDevices []*Device    `json:"support_devices"`
}

// PluginPath 插件的地址
func (p Plugin) PluginPath() string {
	return url.ConcatPath("api", "plugin", p.Name)
}

// LogoURLWithRequest 插件logo的地址
func (p Plugin) LogoURLWithRequest(req *http.Request) string {
	path := url.ConcatPath(p.PluginPath(), p.LogoURL)
	return url.BuildURL(path, nil, req)
}

type DiscoverResponse struct {
	Name         string `json:"name"`
	Identity     string `json:"identity"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	PluginID     string `json:"plugin_id"`
}

type DeviceState map[string]interface{}

type OnDeviceStateChange func(deviceID string, instanceID int, newDS server.Attribute)

//
// func LoadPluginServers() {
//
//	// 判断插件是否已安装，获取已安装的镜像信息
//	// 运行所有镜像的容器
//	plugins, err := GetPlugins()
//	if err != nil {
//		panic(err)
//	}
//	for _, plg := range plugins {
//		docker.ContainerStopByImage(plg.Image.RefStr())
//		plg.Load()
//	}
// }
