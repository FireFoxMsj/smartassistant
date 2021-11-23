package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"

	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

var NotExistErr = errors.New("plugin not exist")

const (
	healthCheckInternal = time.Second * 10 // 存活检查间隔
	offlineTimeout      = time.Second * 15 // 超过这个时间没有响应在线判断为离线
)

type client struct {
	mu      sync.Mutex // clients 锁
	clients map[string]*pluginClient

	devicesCancel        sync.Map
	stateChangeCallbacks []OnDeviceStateChange
}

func (c *client) DeviceConfigs() (configs []DeviceConfig) {

	for _, cli := range c.clients {
		for _, d := range cli.PluginConf.SupportDevices {
			conf := DeviceConfig{
				Device:   *d,
				PluginID: cli.PluginConf.ID,
			}
			configs = append(configs, conf)
		}
	}
	return
}

func (c *client) DeviceConfig(d entity.Device) (config DeviceConfig) {
	if d.Model == types.SaModel {
		return
	}

	cli, err := c.get(d.PluginID)
	if err != nil {
		return
	}

	for _, sd := range cli.PluginConf.SupportDevices {
		if d.Model != sd.Model {
			continue
		}
		config.Device = *sd
		config.PluginID = cli.PluginConf.ID
		return
	}
	return
}

func (c *client) Disconnect(identity, pluginID string, authParams map[string]string) (err error) {
	req := proto.AuthReq{
		Identity: identity,
		Params:   authParams,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	cli, err := c.get(pluginID)
	if err != nil {
		return
	}
	_, err = cli.protoClient.Disconnect(ctx, &req)
	if err != nil {
		return
	}

	v, loaded := c.devicesCancel.LoadAndDelete(identity)
	if loaded {
		if cancel, ok := v.(context.CancelFunc); ok {
			cancel()
		}
	}
	return nil
}

func NewClient(callbacks ...OnDeviceStateChange) *client {
	return &client{
		clients:              make(map[string]*pluginClient),
		stateChangeCallbacks: callbacks,
	}
}

func (c *client) get(domain string) (*pluginClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cli, ok := c.clients[domain]
	if ok {
		return cli, nil
	}
	return nil, NotExistErr
}

func (c *client) Add(cli *pluginClient) {

	c.mu.Lock()
	c.clients[cli.pluginID] = cli
	c.mu.Unlock()
	go c.ListenStateChange(cli.pluginID)
	// 查找该插件所有的设备
	devices, err := entity.GetDevicesByPluginID(cli.pluginID)
	if err != nil {
		return
	}
	for _, device := range devices {
		// 监听所有设备是否在线
		go func(d entity.Device) {
			if err := c.HealthCheck(d); err != nil {
				logger.Error(err)
			}
		}(device)
	}
}

func (c *client) Remove(pluginID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	cli, ok := c.clients[pluginID]
	if ok {
		delete(c.clients, pluginID)
		go cli.Stop()
	}
	return nil
}

// DevicesDiscover 发现设备，并且通过 channel 返回给调用者
func (c *client) DevicesDiscover(ctx context.Context) <-chan DiscoverResponse {
	out := make(chan DiscoverResponse, 1)
	go func() {
		var wg sync.WaitGroup
		for _, cli := range c.clients {
			wg.Add(1)
			go func(cli *pluginClient) {
				defer wg.Done()
				logger.Debug("listening plugin Discovering...")
				cli.DeviceDiscover(ctx, out)
				logger.Debug("plugin listening done")
			}(cli)
		}
		wg.Wait()
		close(out)
	}()
	return out
}

func (c *client) ListenStateChange(pluginID string) {
	cli, err := c.get(pluginID)
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	cli.cancel = cancel
	pdc, err := cli.protoClient.StateChange(ctx, &proto.Empty{})
	if err != nil {
		logger.Error("state onDeviceStateChange error:", err)
		return
	}
	logger.Println("StateChange recv...")
	for {
		resp, err := pdc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Println(err)
			// TODO retry
			break
		}
		logger.Debugf("get state onDeviceStateChange resp: %s,%d,%s\n",
			resp.Identity, resp.InstanceId, string(resp.Attributes))
		var attr server.Attribute
		_ = json.Unmarshal(resp.Attributes, &attr)
		d, err := entity.GetPluginDevice(cli.areaID, cli.pluginID, resp.Identity)
		if err != nil {
			logger.Errorf("ListenStateChange error:%s", err.Error())
			continue
		}

		for _, callback := range c.stateChangeCallbacks {
			a := entity.Attribute{
				Attribute:  attr,
				InstanceID: int(resp.InstanceId),
			}
			go func(cb OnDeviceStateChange) {
				if err := cb(d, a); err != nil {
					logger.Errorf("state change callback err: %s", err.Error())
				}
			}(callback)
		}
	}
	logger.Println("StateChangeFromPlugin exit")
}

func (c *client) SetAttributes(d entity.Device, data json.RawMessage) (result []byte, err error) {
	req := proto.SetAttributesReq{
		Identity: d.Identity,
		Data:     data,
	}
	logger.Debug("set attributes: ", string(data))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	cli, err := c.get(d.PluginID)
	if err != nil {
		return
	}
	_, err = cli.protoClient.SetAttributes(ctx, &req)
	if err != nil {
		logger.Error(err)
		return
	}
	return
}

// Connect 连接设备
func (c *client) Connect(identity, pluginID string, authParams map[string]string) (das DeviceAttributes, err error) {
	req := proto.AuthReq{
		Identity: identity,
		Params:   authParams,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5) // 配对认证过程较长可能长达数分钟
	defer cancel()

	cli, err := c.get(pluginID)
	if err != nil {
		return
	}
	resp, err := cli.protoClient.Connect(ctx, &req)
	if err != nil {
		return
	}
	logger.Debugf("connect resp: %#v\n", resp)
	das = parseAttrsResp(identity, resp)
	return
}

func parseAttrsResp(identity string, resp *proto.GetAttributesResp) DeviceAttributes {

	var instances []Instance
	for _, instance := range resp.Instances {
		var attrs []Attribute
		_ = json.Unmarshal(instance.Attributes, &attrs)
		i := Instance{
			Type:       instance.Type,
			InstanceId: int(instance.InstanceId),
			Attributes: attrs,
		}
		instances = append(instances, i)
	}
	return DeviceAttributes{
		Identity:  identity,
		Instances: instances,
	}
}

func (c *client) GetAttributes(d entity.Device) (das DeviceAttributes, err error) {
	req := proto.GetAttributesReq{Identity: d.Identity}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	cli, err := c.get(d.PluginID)
	if err != nil {
		return
	}
	resp, err := cli.protoClient.GetAttributes(ctx, &req)
	if err != nil {
		return
	}
	logger.Debugf("GetAttributes resp: %#v\n", resp)
	das = parseAttrsResp(d.Identity, resp)
	return
}

func (c *client) HealthCheck(d entity.Device) (err error) {
	cli, err := c.get(d.PluginID)
	if err != nil {
		return
	}
	return cli.HealthCheck(d.Identity)
}

func (c *client) IsOnline(d entity.Device) bool {
	cli, err := c.get(d.PluginID)
	if err != nil {
		return false
	}
	return cli.IsOnline(d.Identity)
}

type pluginClient struct {
	areaID      uint64
	pluginID    string
	protoClient proto.PluginClient // 请求插件服务的grpc客户端
	cancel      context.CancelFunc
	ctx         context.Context
	PluginConf  Plugin

	deviceLastOnlineTime sync.Map
}

func newClient(areaID uint64, plgID, key string, plgConf Plugin) (*pluginClient, error) {
	cli, err := clientv3.NewFromURL(etcdURL)
	if err != nil {
		return nil, err
	}
	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(fmt.Sprintf("etcd:///%s", key), grpc.WithInsecure(), grpc.WithResolvers(etcdResolver))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &pluginClient{
		areaID:      areaID,
		pluginID:    plgID,
		protoClient: proto.NewPluginClient(conn),
		ctx:         ctx,
		cancel:      cancel,
		PluginConf:  plgConf,
	}, nil
}

func (pc *pluginClient) Stop() {
	if pc.cancel != nil {
		pc.cancel()
	}
}

func (pc *pluginClient) DeviceDiscover(ctx context.Context, out chan<- DiscoverResponse) {

	pdc, err := pc.protoClient.Discover(ctx, &proto.Empty{})
	if err != nil {
		logger.Warning(err)
		return
	}
	for {
		select {
		case <-pc.ctx.Done():
			return
		case <-ctx.Done():
			return
		default:
			resp, err := pdc.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				logger.Error(err)
				continue
			}
			// 设备是否在线
			if err = pc.healthCheck(resp.Identity); err != nil {
				logger.Error(err)
				continue
			}
			if !pc.IsOnline(resp.Identity) {
				continue
			}

			device := DiscoverResponse{
				Identity:     resp.Identity,
				Model:        resp.Model,
				Manufacturer: resp.Manufacturer,
				Name:         fmt.Sprintf("%s_%s", resp.Manufacturer, resp.Model),
				PluginID:     pc.pluginID,
				AuthRequired: resp.AuthRequired,
			}
			out <- device
		}
	}
}

// HealthCheck 监听设备的在线状态
func (pc *pluginClient) HealthCheck(identity string) error {

	_, loaded := pc.deviceLastOnlineTime.Load(identity)
	if loaded { // 已经监听了则直接返回,防止重复
		return nil
	}
	logger.Debugf("%s start health check...", identity)
	ticker := time.NewTicker(healthCheckInternal)
	for {
		select {
		case <-pc.ctx.Done():
			logger.Debugf("%s health check done", identity)
			return nil
		case <-ticker.C:
			if err := pc.healthCheck(identity); err != nil {
				logger.Errorf("%s HealthCheck err: %s", identity, err.Error())
			}
		}
	}
}

// healthCheck 查看设备的在线状态
func (pc *pluginClient) healthCheck(identity string) error {

	req := proto.HealthCheckReq{Identity: identity}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := pc.protoClient.HealthCheck(ctx, &req)
	if err != nil {
		return err
	}
	if resp.Online {
		pc.deviceLastOnlineTime.Store(identity, time.Now())
	}
	return nil
}

func (pc *pluginClient) IsOnline(identity string) bool {
	if v, ok := pc.deviceLastOnlineTime.Load(identity); ok {
		lastOnlineTime := v.(time.Time)
		return lastOnlineTime.Add(offlineTimeout).After(time.Now())
	}
	return false
}

// GetPluginConfig 获取插件配置
func GetPluginConfig(addr, pluginID string) (config Plugin, err error) {
	url := fmt.Sprintf("http://%s/api/plugin/%s/config.json", addr, pluginID)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if err = json.Unmarshal(data, &config); err != nil {
		return
	}
	return
}
