package plugin

import (
	errors2 "errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/go-playground/validator/v10"
	"github.com/zhiting-tech/smartassistant/modules/config"
	version2 "github.com/zhiting-tech/smartassistant/modules/utils/version"
	"os"
	"path/filepath"
	"strings"

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
	Model string     `json:"model" yaml:"model"`
	Name  string     `json:"name" yaml:"name"`
	Type  DeviceType `json:"type" yaml:"type"` // 设备类型

	Logo         string `json:"logo" yaml:"logo"`                 // 设备logo相对路径
	Control      string `json:"control" yaml:"control"`           // 设备控制页面相对路径
	Provisioning string `json:"provisioning" yaml:"provisioning"` // 设备置网页面相对路径
}

type PluginConfig struct {
	Name           string       `json:"name" validate:"required"`            // 插件名称
	Version        string       `json:"version" validate:"required"`         // 版本
	Info           string       `json:"info"`                                // 介绍
	SupportDevices []DeviceInfo `json:"support_devices" validate:"required"` // 支持的设备
}

type DeviceInfo struct {
	Model string `json:"model" validate:"required"`
	Name  string `json:"name" validate:"required"`

	Logo         string `json:"logo" validate:"required"`    // 设备logo相对路径
	Control      string `json:"control" validate:"required"` // 设备控制页面相对路径
	Provisioning string `json:"provisioning"`                // 设备置网页面相对路径
}

// ID 根据配置生成插件ID
func (p PluginConfig) ID() string {
	return p.Name
}
func (p PluginConfig) Validate() error {
	defaultValidator := validator.New()
	defaultValidator.SetTagName("validate")
	return defaultValidator.Struct(p)
}

type Plugin struct {
	ID             string    `json:"id" yaml:"id"`
	Name           string    `json:"name" yaml:"name"`
	Image          string    `json:"image" yaml:"image"`
	Version        string    `json:"version" yaml:"version"`
	Brand          string    `json:"brand" yaml:"brand"`
	Info           string    `json:"info" yaml:"info"`
	DownloadURL    string    `json:"download_url" yaml:"download_url"` // 插件静态文件下载地址
	SupportDevices []*Device `json:"support_devices" yaml:"support_devices"`
	Source         string    `json:"source" yaml:"source"` // 插件来源
	AreaID         uint64    `json:"area_id" yaml:"area_id"`
}

func NewFromEntity(p entity.PluginInfo) Plugin {
	return Plugin{
		ID:      p.PluginID,
		Name:    p.PluginID,
		Image:   p.Image,
		Version: p.Version,
		Info:    p.Info,
		AreaID:  p.AreaID,
		Source:  p.Source,
	}
}

// IsDevelopment 是否开发者上传的插件
func (p Plugin) IsDevelopment() bool {
	return p.Source == entity.SourceTypeDevelopment
}

func (p Plugin) IsAdded() bool {
	// return docker.GetClient().IsImageAdd(p.Image.RefStr())
	return entity.IsPluginAdd(p.ID, p.AreaID)
}
func (p Plugin) IsNewest() bool {
	if p.Source == entity.SourceTypeDevelopment {
		return true
	}
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
	isRunning, _ := docker.GetClient().ContainerIsRunningByImage(p.Image)
	return isRunning
}

// Up 启动插件
func (p Plugin) Up() (err error) {
	logger.Info("up plugin:", p.Name)
	_, err = RunPlugin(p)
	if err != nil && strings.Contains(err.Error(), "already in use") {
		return nil
	}
	return err
}
func (p Plugin) UpdateOrInstall() (err error) {
	if p.IsAdded() {
		return p.Update()
	}
	return p.Install()
}

// Install 安装并且启动插件
func (p Plugin) Install() (err error) {

	// TODO 镜像没build或者build失败则不能安装

	if !p.IsDevelopment() {
		if err = docker.GetClient().Pull(p.Image); err != nil {
			return
		}
	}
	if err = p.Up(); err != nil {
		return
	}

	var pi = entity.PluginInfo{
		AreaID:   p.AreaID,
		PluginID: p.ID,
		Image:    p.Image,
		Info:     p.Info,
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
	if p.Source == entity.SourceTypeDevelopment {
		return errors2.New("plugin in development can't update")
	}
	logger.Info("update plugin:", p.ID)
	if err = docker.GetClient().ContainerStopByImage(p.Image); err != nil {
		logger.Error(err.Error())
	}
	if err = docker.GetClient().ImageRemove(p.Image); err != nil {
		logger.Error(err.Error())
	}
	return p.Install()
}

// Remove 删除插件
func (p Plugin) Remove() (err error) {
	logger.Info("removing plugin", p.ID)
	if err = docker.GetClient().ContainerStopByImage(p.Image); err != nil {
		logger.Error(err.Error())
	}

	if err = docker.GetClient().ImageRemove(p.Image); err != nil {
		logger.Error(err.Error())
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

func GetInfoFromDeviceAttrs(pluginID string, das DeviceAttributes) (d entity.Device, err error) {
	d.Identity = das.Identity
	d.PluginID = pluginID
	for _, ins := range das.Instances {
		if ins.Type == "info" {
			for _, attr := range ins.Attributes {
				switch attr.Attribute.Attribute {
				case "model":
					d.Model = attr.Val.(string)
					d.Name = attr.Val.(string)
				case "manufacturer":
					d.Manufacturer = attr.Val.(string)
				}
			}
			return
		}
	}
	err = errors2.New("no instance info found")
	return
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
	AuthRequired bool   `json:"auth_required"`
}

// RunPlugin 运行插件
func RunPlugin(plg Plugin) (containerID string, err error) {
	conf := container.Config{
		Image: plg.Image,
		Env:   []string{fmt.Sprintf("PLUGIN_DOMAIN=%s", plg.ID)},
	}
	// 映射插件目录到宿主机上
	source := filepath.Join(config.GetConf().SmartAssistant.HostRuntimePath,
		"data", "plugin", plg.Brand, plg.Name)
	if err = os.MkdirAll(source, os.ModePerm); err != nil {
		return
	}
	target := "/app/data/"
	logger.Debugf("mount %s to %s", source, target)

	hostConf := container.HostConfig{
		NetworkMode: "host",
		AutoRemove:  true,
		Mounts: []mount.Mount{
			{Type: mount.TypeBind, Source: source, Target: target},
		},
	}
	if config.GetConf().SmartAssistant.FluentdAddress != "" {
		//设置容器的logging, driver
		hostConf.LogConfig = container.LogConfig{
			Type: "fluentd",
			Config: map[string]string{
				"fluentd-address": config.GetConf().SmartAssistant.FluentdAddress,
				"tag":             fmt.Sprintf("smartassistant.plugin.%s", plg.Image),
			},
		}
	}
	return docker.GetClient().ContainerRun(plg.Image, conf, hostConf)
}
