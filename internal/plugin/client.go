package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

type Client struct {
	pluginID    string
	protoClient proto.PluginService
	cancel      context.CancelFunc
}

func newClient(plgID string, api proto.PluginService) *Client {
	return &Client{
		pluginID:    plgID,
		protoClient: api,
	}
}

func (c *Client) Run(cb OnDeviceStateChange) {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	pdc, err := c.protoClient.StateChange(ctx, &proto.Empty{})
	if err != nil {
		logrus.Error("state change error:", err)
		return
	}
	log.Println("StateChange done,recv...")
	for {
		resp, err := pdc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			// TODO retry
			break
		}
		var attr server.Attribute
		logrus.Debugf("get state change resp: %s,%d,%s\n",
			resp.Identity, resp.InstanceId, string(resp.Attributes))
		json.Unmarshal(resp.Attributes, &attr)
		// s将设备identity转换为数据库设备id
		d, err := entity.GetDeviceByIdentity(resp.Identity)
		if err != nil {
			log.Println(err)
			continue
		}
		go cb(d.Identity, int(resp.InstanceId), attr)
	}
	log.Println("StateChangeFromPlugin exit")
}

func (c *Client) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Client) DeviceDiscover(ctx context.Context, out chan<- DiscoverResponse) {
	pdc, err := c.protoClient.Discover(ctx, &proto.Empty{})
	if err != nil {
		logrus.Warning(err)
		return
	}
	for {
		resp, err := pdc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logrus.Warning(err)
			continue
		}
		device := DiscoverResponse{
			Identity:     resp.Identity,
			Model:        resp.Model,
			Manufacturer: resp.Manufacturer,
			Name:         fmt.Sprintf("%s_%s_%s", resp.Manufacturer, resp.Model, resp.Identity),
			PluginID:     c.pluginID,
		}
		out <- device
	}
}

func (c *Client) SetAttributes(identity string, data json.RawMessage) (result []byte, err error) {
	req := proto.SetAttributesReq{
		Identity: identity,
		Data:     data,
	}
	logrus.Debug("set attributes: ", string(data))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err = c.protoClient.SetAttributes(ctx, &req)
	if err != nil {
		logrus.Error(err)
		return
	}
	return
}

func (c *Client) GetAttributes(identity string) (*proto.GetAttributesResp, error) {
	req := proto.GetAttributesReq{Identity: identity}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := c.protoClient.GetAttributes(ctx, &req)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("state resp: %#v\n", resp)
	return resp, nil
}

func SetAttributes(plugin, identity string, data json.RawMessage) (err error) {
	plgCli, err := GetManager().ClientGet(plugin)
	if err != nil {
		return
	}
	_, err = plgCli.SetAttributes(identity, data)
	return
}

// GetControlAttributes 获取设备属性（不包括设备型号、厂商等属性）
func GetControlAttributes(d entity.Device) (attributes []entity.Attribute, err error) {
	das, err := GetDeviceAttributes(d.PluginID, d.Identity)
	if err != nil {
		return
	}
	for _, instance := range das.Instances {
		if instance.Type == "info" {
			continue
		}
		as := GetInstanceControlAttributes(instance)
		attributes = append(attributes, as...)
	}
	return
}

// GetInstanceControlAttributes 获取实例的控制属性
func GetInstanceControlAttributes(instance Instance) (attributes []entity.Attribute) {
	for _, attr := range instance.Attributes {

		// 仅返回能控制的属性
		if attr.Attribute.Attribute == "name" {
			continue
		}
		a := entity.Attribute{
			Attribute:  attr.Attribute,
			InstanceID: instance.InstanceId,
		}
		attributes = append(attributes, a)
	}
	return
}
func GetControlAttributeByID(d entity.Device, instanceID int, attr string) (attribute entity.Attribute, err error) {
	as, err := GetControlAttributes(d)
	if err != nil {
		return
	}

	for _, a := range as {
		if a.InstanceID == instanceID && a.Attribute.Attribute == attr {
			return a, nil
		}
	}
	err = fmt.Errorf("plugin %s d %s instance id %s attr  %s not found",
		d.PluginID, d.Identity, instanceID, attr)
	return
}

func GetUserDeviceAttributes(userID int, plugin, identity string) (das DeviceAttributes, err error) {
	das, err = GetDeviceAttributes(plugin, identity)
	if err != nil {
		return
	}
	device, err := entity.GetDeviceByIdentity(identity)
	if err != nil {
		return
	}
	for i, instance := range das.Instances {
		for j, a := range instance.Attributes {
			if entity.IsDeviceControlPermitByAttr(userID, device.ID,
				instance.InstanceId, a.Attribute.Attribute) {
				das.Instances[i].Attributes[j].CanControl = true
			}
		}
	}
	return
}
func GetDeviceAttributes(plugin, identity string) (das DeviceAttributes, err error) {

	plgCli, err := GetManager().ClientGet(plugin)
	if err != nil {
		return
	}
	var getAttrResp *proto.GetAttributesResp
	getAttrResp, err = plgCli.GetAttributes(identity)
	if err != nil {
		return
	}
	var instances []Instance
	for _, instance := range getAttrResp.Instances {
		var attrs []Attribute
		json.Unmarshal(instance.Attributes, &attrs)
		i := Instance{
			Type:       instance.Type,
			InstanceId: int(instance.InstanceId),
			Attributes: attrs,
		}
		instances = append(instances, i)
	}
	das = DeviceAttributes{
		Identity:  identity,
		Instances: instances,
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
	Type      string     `json:"type"`
	Instances []Instance `json:"instances"`
}
