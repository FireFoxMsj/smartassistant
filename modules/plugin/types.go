package plugin

import (
	"encoding/json"
	errors2 "errors"
	version2 "github.com/zhiting-tech/smartassistant/modules/utils/version"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

type DeviceType string

const (
	TypeLight          DeviceType = "light"           // 灯
	TypeSwitch         DeviceType = "switch"          // 开关
	TypeOutlet         DeviceType = "outlet"          // 插座
	TypeRoutingGateway DeviceType = "routing_gateway" // 路由网关
	TypeSecurity       DeviceType = "security"        // 安防
)

type Device struct {
	Model string     `json:"model"`
	Name  string     `json:"name"`
	Type  DeviceType `json:"type"` // 设备类型

	Logo         string `json:"logo" yaml:"logo"`                 // 设备logo相对路径
	Control      string `json:"control" yaml:"control"`           // 设备控制页面相对路径
	Provisioning string `json:"provisioning" yaml:"provisioning"` // 设备置网页面相对路径
}

type Plugin struct {
	ID             string       `json:"id"`
	Image          docker.Image `json:"image"`
	LogoURL        string       `json:"logo_url"`
	Version        string       `json:"version"`
	Brand          string       `json:"brand"`
	Info           string       `json:"info"`
	DownloadURL    string       `json:"download_url"` // 插件静态文件下载地址
	SupportDevices []*Device    `json:"support_devices"`
	Source         string       `json:"source"` // 插件来源
	AreaID         uint64       `json:"area_id"`
}

// IsDevelopment 是否开发者上传的插件
func (p Plugin) IsDevelopment() bool {
	return p.Source == entity.SourceTypeDevelopment
}

// BrandLogoURL 插件logo的地址
func (p Plugin) BrandLogoURL(req *http.Request) string {
	plg, err := GetGlobalManager().Get(p.ID)
	if err != nil {
		logrus.Error(err)
		return ""
	}
	// TODO 改为返回插件中的图片地址
	return plg.LogoURL
}
func (p Plugin) IsAdded() bool {
	// return docker.GetClient().IsImageAdd(p.Image.RefStr())
	return entity.IsPluginAdd(p.ID, p.AreaID)
}
func (p Plugin) IsNewest() bool {
	return false // 方便开发更新插件

	pluginInfo, err := entity.GetPlugin(p.ID, p.AreaID)
	if err != nil {
		logger.Errorf("get plugin info fail: %v\n", err)
		return true
	}
	greater, err := version2.Greater(p.Version, pluginInfo.Version)
	if err != nil {
		logger.Errorf("compare plugin version fail: %v\n", err)
		return true
	}
	return greater
}

func (p Plugin) IsRunning() bool {
	isRunning, _ := docker.GetClient().ContainerIsRunningByImage(p.Image.RefStr())
	return isRunning
}

// Up 启动插件
func (p Plugin) Up() (err error) {
	logger.Info("up plugin:", p.Image.Name)
	_, err = docker.GetClient().ContainerRunByImage(p.Image)
	if err != nil && strings.Contains(err.Error(), "already in use") {
		return nil
	}
	return err
}

// Install 安装并且启动插件
func (p Plugin) Install() (err error) {

	// TODO 镜像没build或者build失败则不能安装

	if !p.IsDevelopment() {
		if err = docker.GetClient().Pull(p.Image.RefStr()); err != nil {
			return
		}
	}
	if err = p.Up(); err != nil {
		return
	}

	// 默认插件，安装时进行插入记录
	info, _ := json.Marshal(p) // TODO 读取配置
	var pi = entity.PluginInfo{
		AreaID:   p.AreaID,
		PluginID: p.ID,
		Info:     info,
		Status:   entity.StatusInstallSuccess,
		Version:  p.Version,
		Source:   p.Source,
		Brand:    p.Brand,
	}
	if err = entity.SavePluginInfo(pi); err != nil {
		logger.Errorf("UpdatePluginStatus err: %s", err.Error())
		return
	}
	return
}

// Update 更新插件
func (p Plugin) Update() (err error) {
	logger.Info("update plugin:", p.ID)
	if err = docker.GetClient().ContainerStopByImage(p.Image.RefStr()); err != nil {
		return
	}
	// TODO 开发者插件更新暂时不删除镜像：因为开发者的插件包build完就删除了，删掉镜像后面就用不了
	if p.Source == entity.SourceTypeDefault {
		if err = docker.GetClient().ImageRemove(p.Image.RefStr()); err != nil {
			return
		}
	}
	return p.Install()
}

// Remove 删除插件
func (p Plugin) Remove() (err error) {
	logger.Info("removing plugin", p.ID)
	if err = docker.GetClient().ContainerStopByImage(p.Image.RefStr()); err != nil {
		return
	}
	if err = docker.GetClient().ImageRemove(p.Image.RefStr()); err != nil {
		return
	}
	if err = entity.DelDevicesByPlgID(p.ID); err != nil {
		return
	}
	if err = entity.DelPlugin(p.ID, p.AreaID); err != nil {
		return
	}
	return
}

type Attribute struct {
	server.Attribute
	CanControl bool `json:"can_control"`
}

type Instance struct {
	Type       string      `json:"type"`
	InstanceId int         `json:"instance_id"`
	Attributes []Attribute `json:"attributes"`
}

type DeviceAttributes struct {
	Identity  string     `json:"identity"`
	Instances []Instance `json:"instances"`
	Online    bool       `json:"online"`
}

type OnDeviceStateChange func(d entity.Device, attr entity.Attribute) error

func DefaultOnDeviceStateChange(d entity.Device, attr entity.Attribute) error {
	return errors2.New("OnDeviceStateChange not implement")
}

type DiscoverResponse struct {
	Name         string `json:"name"`
	Identity     string `json:"identity"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	PluginID     string `json:"plugin_id"`
}
