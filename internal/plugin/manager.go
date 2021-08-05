package plugin

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/transport"
	"github.com/zhiting-tech/smartassistant/internal/config"
	"github.com/zhiting-tech/smartassistant/internal/plugin/docker"
	"github.com/zhiting-tech/smartassistant/internal/types"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
	plugin2 "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"

	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/pkg/reverseproxy"
)

var (
	manager     *Manager
	managerOnce sync.Once
)

// Manager 插件服务器，管理插件，转发插件消息
type Manager struct {
	stateChangeCB OnDeviceStateChange // 接收到设备状态变化后会调用此回调
	docker        *docker.Client
	mu            sync.Mutex // clients 锁
	clients       map[string]*Client
	plugins       map[string]*Plugin
}

func defaultStateChangeCB(identity string, instanceID int, newDS plugin2.Attribute) {
	logrus.Warning("state change default handler")
}

func GetManager() *Manager {
	managerOnce.Do(func() {
		manager = &Manager{
			stateChangeCB: defaultStateChangeCB,
			docker:        docker.GetClient(),
			clients:       make(map[string]*Client),
			plugins:       make(map[string]*Plugin),
		}
	})
	return manager
}

// Run 启动服务，扫描插件并且连接通讯
func (m *Manager) Run(ctx context.Context) {
	logrus.Info("starting plugin manager")
	// 加载插件列表
	if err := m.LoadPlugins(); err != nil {
		panic(err)
	}
	// 发现运行中的插件，并且启动 client
	go m.Discovery()

	// 等待其他容器启动，判断如果插件没有运行，则启动
	time.Sleep(5 * time.Second)
	go m.StartPlugins()

	// TODO 扫描已安装的插件，并且启动，连接 state change...
	<-ctx.Done()
	// TODO 断开连接
	logrus.Warning("plugin manager stopped")
}

// SetStateChangeCB 设置
func (m *Manager) SetStateChangeCB(cb OnDeviceStateChange) {
	m.stateChangeCB = cb
}

// PluginInstall 安装并且启动插件
func (m *Manager) PluginInstall(id string) (err error) {
	p, err := m.GetPlugin(id)
	if err != nil {
		logrus.Warning("plugin not found", id)
		return
	}
	logrus.Info("loading plugin ", p.Name)
	if err = m.docker.Pull(p.Image.RefStr()); err != nil {
		return
	}
	// TODO 接口有延迟
	time.Sleep(1 * time.Second)
	_, err = m.docker.ContainerRunByImage(p.Image)
	if err != nil && strings.Contains(err.Error(), "already in use") {
		return nil
	}
	return
}

// PluginUpdate 更新插件 TODO 优雅一点
func (m *Manager) PluginUpdate(id string) (err error) {
	p, err := m.GetPlugin(id)
	if err != nil {
		return
	}
	logrus.Info("update plugin", id)
	if err = m.docker.ContainerStopByImage(p.Image.RefStr()); err != nil {
		return
	}
	return m.PluginInstall(id)
}

// PluginRemove 删除插件
func (m *Manager) PluginRemove(id string) (err error) {
	p, err := m.GetPlugin(id)
	if err != nil {
		return
	}
	logrus.Info("removing plugin", id)
	if err = m.docker.ContainerStopByImage(p.Image.RefStr()); err != nil {
		return
	}
	if err = entity.DelDevicesByPlgID(id); err != nil {
		return
	}
	return
}

func (m *Manager) PluginStatus(id string) (isAdded, isNewest bool) {
	p, err := m.GetPlugin(id)
	if err != nil {
		return
	}
	isNewest, _ = m.docker.IsImageNewest()
	isAdded = m.docker.IsImageAdd(p.Image.RefStr())
	return
}

// DeviceDiscover 发现 plugins 下设备，并且通过 channel 返回给调用者
func (m *Manager) DeviceDiscover(ctx context.Context) <-chan DiscoverResponse {
	out := make(chan DiscoverResponse, 1)
	go func() {
		var wg sync.WaitGroup
		for _, cli := range m.clients {
			wg.Add(1)
			go func(c *Client) {
				defer wg.Done()
				logrus.Debug("listening plugin Discovering...")
				c.DeviceDiscover(ctx, out)
				logrus.Debug("plugin listening done")
			}(cli)
		}
		wg.Wait()
		close(out)
	}()
	return out
}

func (m *Manager) registerService(service *registry.Service) error {

	if service == nil || len(service.Nodes) == 0 {
		return nil
	}
	domain := service.Name
	serviceAddr := service.Nodes[0].Address
	logrus.Debugf("register service %s:%sS from etcd", service.Name, serviceAddr)
	if strings.HasSuffix(domain, "http") {
		return reverseproxy.RegisterUpstream(strings.TrimSuffix(domain, ".http"), serviceAddr)
	} else {
		return m.ClientAdd(domain)
	}

}
func (m *Manager) unregisterService(service *registry.Service) error {

	if service == nil || len(service.Nodes) == 0 {
		return nil
	}
	logrus.Debugf("unregister service %s from etcd", service.Name)
	domain := service.Name
	if strings.HasSuffix(domain, "http") {
		return reverseproxy.UnregisterUpstream(strings.TrimSuffix(domain, ".http"))
	} else {
		return m.ClientRemove(domain)
	}
}

// Discovery 监听etcd发现服务
func (m *Manager) Discovery() {
	log.Println("start discovering service")

	// list and register services from etcd
	services, err := DefaultRegistry().ListServices()
	if err != nil {
		logrus.Error(err)
		return
	}
	for _, service := range services {
		if err = m.registerService(service); err != nil {
			logrus.Error(err)
		}
	}

	// watch etcd service change
	w, err := DefaultRegistry().Watch()
	if err != nil {
		panic(err)
	}
	for {
		var r *registry.Result
		if r, err = w.Next(); err != nil {
			log.Println(err)
			continue
		}
		if r.Action == "delete" {
			if err = m.unregisterService(r.Service); err != nil {
				logrus.Error(err)
			}
		} else {
			if err = m.registerService(r.Service); err != nil {
				logrus.Error(err)
			}
		}
	}
}

// LoadPlugins 加载插件列表
func (m *Manager) LoadPlugins() (err error) {
	plgsFile, err := os.Open("plugins.json")
	if err != nil {
		return
	}
	defer plgsFile.Close()

	data, err := ioutil.ReadAll(plgsFile)
	if err != nil {
		return
	}
	plgs := make([]Plugin, 0)
	if err = json.Unmarshal(data, &plgs); err != nil {
		return
	}
	for i, plg := range plgs {
		logrus.Println("load plugin:", plg.ID, plg)
		m.plugins[plg.ID] = &plgs[i]
	}
	return
}

func (m *Manager) DevicePluginURL(device entity.Device, req *http.Request, token string) string {
	if device.Model == types.SaModel {
		return ""
	}

	q := map[string]interface{}{
		"device_id": device.ID,
		"identity":  device.Identity,
		"model":     device.Model,
		"name":      device.Name,
		"token":     token,
		"sa_id":     config.GetConf().SmartAssistant.ID,
		"plugin_id": device.PluginID,
	}
	path := url.ConcatPath(device.PluginPath(), "html/")
	return url.BuildURL(path, q, req)
}

// GetPlugin 获取单个插件信息
func (m *Manager) GetPlugin(id string) (*Plugin, error) {
	if plg, ok := m.plugins[id]; ok {
		return plg, nil
	}
	return nil, errors.New(status.PluginDomainNotExist)
}

// ListPlugin 插件列表
func (m *Manager) ListPlugin() []*Plugin {
	res := make([]*Plugin, 0, len(m.plugins))
	for _, plg := range m.plugins {
		res = append(res, plg)
	}
	return res
}

// StartPlugins 启动所有已安装的插件
func (m *Manager) StartPlugins() {
	for _, plg := range m.plugins {
		if !m.docker.IsImageAdd(plg.Image.RefStr()) {
			continue
		}
		isRunning, _ := m.docker.ContainerIsRunningByImage(plg.Image.RefStr())
		if isRunning {
			continue
		}
		// 如果镜像没运行，则启动
		if err := m.PluginInstall(plg.ID); err != nil {
			logrus.Error("plugin start error:", err)
		}
	}
}

func (m *Manager) ClientGet(domain string) (*Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cli, ok := m.clients[domain]
	if ok {
		return cli, nil
	}
	return nil, NotExistErr
}

func (m *Manager) ClientAdd(domain string) error {
	ps := proto.NewPluginService(domain, DefaultClient())
	cli := newClient(domain, ps)
	m.mu.Lock()
	m.clients[domain] = cli
	m.mu.Unlock()
	go cli.Run(m.stateChangeCB)
	return nil
}

func (m *Manager) ClientRemove(domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cli, ok := m.clients[domain]
	if ok {
		delete(m.clients, domain)
		go cli.Stop()
	}
	return nil
}

func DefaultRegistry() registry.Registry {
	addr := registry.Addrs("etcd:2379")
	timeout := registry.Timeout(10 * time.Second)
	return etcd.NewRegistry(addr, timeout)
}

func DefaultClient() client.Client {
	return grpc.NewClient(
		client.Registry(DefaultRegistry()),
		client.DialTimeout(time.Minute),
		client.RequestTimeout(time.Minute),
		client.Transport(transport.NewTransport()),
	)
}
