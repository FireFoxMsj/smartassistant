package plugin

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type Manager interface {
	// Load 加载并返回所有插件
	Load() (map[string]*Plugin, error)
	Get(id string) (*Plugin, error)
}

type Client interface {
	DevicesDiscover(ctx context.Context) <-chan DiscoverResponse
	GetAttributes(device entity.Device) (DeviceAttributes, error)
	SetAttributes(device entity.Device, data json.RawMessage) (result []byte, err error)
	HealthCheck(entity.Device) error
	IsOnline(entity.Device) bool
	Disconnect(entity.Device) error
	DeviceInfo(entity.Device) Info
}

type Info struct {
	Logo         string `json:"logo" yaml:"control"`              // 设备logo地址相对路径
	Control      string `json:"control" yaml:"control"`           // 设备控制页面相对路径
	Provisioning string `json:"provisioning" yaml:"provisioning"` // 设备置网页面相对路径
	Compress     string `json:"compress" yaml:"compress"`         // 压缩包地址
}

var (
	globalManager     Manager
	globalManagerOnce sync.Once

	globalClient     Client
	globalClientOnce sync.Once
)

func SetGlobalClient(c Client) {
	globalClientOnce.Do(func() {
		globalClient = c
	})
}

func GetGlobalClient() Client {
	globalClientOnce.Do(func() {
		globalClient = NewClient(DefaultOnDeviceStateChange)
	})
	return globalClient
}

func GetGlobalManager() Manager {
	globalManagerOnce.Do(func() {
		globalManager = NewManager()
		loadAndUpPlugins(globalManager)
	})
	return globalManager
}

func SetGlobalManager(m Manager) {
	globalManagerOnce.Do(func() {
		globalManager = m
		loadAndUpPlugins(globalManager)
	})
}

// loadAndUpPlugins 加载插件并启动已安装的插件
func loadAndUpPlugins(m Manager) {

	logger.Info("starting plugin globalManager")
	// 加载插件列表
	plugins, err := m.Load()
	if err != nil {
		return
	}
	// 扫描已安装的插件，并且启动，连接 state change...
	// 等待其他容器启动，判断如果插件没有运行，则启动
	time.Sleep(5 * time.Second)
	for _, plg := range plugins {
		if !plg.IsAdded() || plg.IsRunning() {
			continue
		}
		// 如果镜像没运行，则启动
		if upErr := plg.Up(); upErr != nil {
			logger.Error("plugin up error:", upErr)
		}
	}
}
