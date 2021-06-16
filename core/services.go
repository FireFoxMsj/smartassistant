package core

import (
	"errors"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"gitlab.yctc.tech/root/smartassistent.git/utils/pubsub"
)

var PluginNotExistErr = errors.New("plugin not exist")

func install(m M) error {
	pluginID := m.ValString("plugin_id")
	plg := plugin.GetPlugin(pluginID)
	if plg.ID != pluginID {
		return PluginNotExistErr
	}
	if err := plugin.Install(plg); err != nil {
		return err
	}

	eventData := M{
		"plugin_name":     plg.Name,
		"call_service_id": m.Get("call_service_id"),
		"client_key":      m.ValString("client_key"),
	}
	Sass.Bus.Fire(EventInstallPlugin, eventData)
	return nil
}
func remove(m M) error {
	if err := removePlugin(m); err != nil {
		return err
	}
	// 删除设备
	if err := orm.DelDevicesByPlgID(m.ValString("plugin_id")); err != nil {
		return err
	}
	return nil
}

func removePlugin(m M) error {
	pluginID := m.ValString("plugin_id")
	plg := plugin.GetPlugin(pluginID)
	if plg.ID != pluginID {
		return PluginNotExistErr
	}
	if err := plugin.Remove(plg); err != nil {
		return err
	}
	return nil
}

func update(m M) error {
	if err := removePlugin(m); err != nil {
		return err
	}

	if err := install(m); err != nil {
		return err
	}
	return nil
}

// 安装完成后，也要加载
func afterInstall(e Event) error {
	plgName := e.Data.ValString("plugin_name")
	if err := plugin.LoadOne(plgName); err != nil {
		return err
	}
	// 推送消息
	response := CallResponse{
		ID:      e.Data.Get("call_service_id").(int),
		Success: true,
	}

	data := M{"client_key": e.Data.ValString("client_key"), "data": response}
	return Sass.Bus.Fire(EventSingleCast, data)
}

// 发现设备后，插件需要Fire EventDeviceDiscovered 事件以保存数据
// TODO 按实际情况调整
func afterDeviceDiscover(e Event) error {

	d, _ := e.Data.Get("device").(map[string]interface{})
	i, _ := d["identity"]
	identity, _ := i.(string)
	_, err := orm.GetDeviceByIdentity(identity)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 推送消息
	response := CallResponse{
		ID:      e.Data.ValInt("call_service_id"),
		Success: true,
	}

	response.AddResult("device", d)
	data := M{"client_key": e.Data.ValString("client_key"), "data": response}
	return Sass.Bus.Fire(EventSingleCast, data)
}

// getActionsFunc 获取并响应用户对设备的控制权限
func getActionsFunc(args M) error {

	deviceID := args.ValInt("device_id")
	userID := args.ValInt("user_id")

	res, err := getActions(userID, deviceID)
	if err != nil {
		return err
	}
	// 推送消息
	response := CallResponse{
		ID:      args.ValInt("call_service_id"),
		Success: true,
	}
	response.AddResult("actions", res)

	data := M{"client_key": args.ValString("client_key"), "data": response}
	return Sass.Bus.Fire(EventSingleCast, data)
}

type Action struct {
	Cmd      string `json:"cmd"`
	Name     string `json:"name"`
	IsPermit bool   `json:"is_permit"`
}

// getActions 获取设备的权限
func getActions(userID, deviceID int) (map[string]Action, error) {

	device, err := orm.GetDeviceByID(deviceID)
	if err != nil {
		return nil, err
	}

	res := make(map[string]Action)
	actions := orm.GetDeviceActions(device)
	for _, a := range actions {
		action := Action{
			Name:     a.Name,
			Cmd:      a.Cmd,
			IsPermit: orm.IsDeviceControlPermit(userID, deviceID, a.Attribute),
		}
		res[a.Cmd] = action
	}
	return res, nil
}

var StatePubsub = pubsub.NewPubsub()
var i uint32 = 0

func GetAutoIncrID() uint32 {
	atomic.AddUint32(&i, 1)
	return i
}

// GetDeviceStateFromResp 发消息获取设备状态
func GetDeviceStateFromResp(device orm.Device) (map[string]interface{}, error) {

	callID := GetAutoIncrID()
	ch := make(chan interface{})
	topic := strconv.FormatUint(uint64(callID), 10)
	StatePubsub.Subscribe(topic, ch)
	defer StatePubsub.UnSubscribe(topic, ch)

	data := M{
		"id":              device.Identity,
		"call_service_id": callID,
		"service_name":    "state",
	}

	if err := Sass.Services.Call(device.Manufacturer, "state", data); err != nil {
		return nil, err
	}

	timeout := time.NewTimer(5 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case <-timeout.C:
			return nil, nil
		case resp := <-ch:
			log.Println("GetDeviceStateFromResp,state:", resp)
			return resp.(map[string]interface{}), nil
		}
	}
}

func GetMapValue(m map[string]interface{}, args ...string) (v interface{}) {

	for _, arg := range args {
		var ok bool
		if v != nil {
			if m, ok = v.(map[string]interface{}); !ok {
				return nil
			}
		}
		if v, ok = m[arg]; !ok {
			return nil
		}
	}
	return v
}

// stateResp 获取状态响应并通知
func stateResp(event Event) error {

	state := GetMapValue(event.Data, "data", "result", "state")
	if state != nil {
		id := GetMapValue(event.Data, "data", "id").(float64)
		StatePubsub.Publish(strconv.Itoa(int(id)), state)
	}
	return nil
}

func init() {
	domain := "plugin"

	Sass.Services.Register(domain, "install", install)
	Sass.Services.Register(domain, "remove", remove)
	Sass.Services.Register(domain, "update", update)
	Sass.Services.Register(domain, "get_actions", getActionsFunc)

	// 事件
	Sass.Bus.Listen(EventInstallPlugin, afterInstall)
	Sass.Bus.Listen(EventDeviceDiscovered, afterDeviceDiscover)
	Sass.Bus.Listen(EventSingleCast, stateResp)

}
