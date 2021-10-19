package websocket

import (
	"context"
	"encoding/json"
	errors2 "errors"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/zhiting-tech/smartassistant/modules/api/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gorm.io/gorm"
)

type client struct {
	key    string
	areaID uint64
	conn   *ws.Conn
	send   chan []byte
	bucket *bucket
}

func (cli *client) Close() error {
	close(cli.send)
	return cli.conn.Close()
}

type ActionWrap struct {
	Cmd      string `json:"cmd"`
	Name     string `json:"name"`
	IsPermit bool   `json:"is_permit"`
}

type DeviceWrap struct {
	Cmd      string `json:"cmd"`
	Name     string `json:"name"`
	IsPermit bool   `json:"is_permit"`
}

var _callServicePool sync.Pool

// 解析 WebSocket 消息，并且调用业务逻辑
func (cli *client) handleWsMessage(data []byte, user *session.User) (err error) {
	cs := _callServicePool.Get().(*callService)
	defer _callServicePool.Put(cs)
	cs.reset()
	if err = json.Unmarshal(data, cs); err != nil {
		return
	}

	logger.Printf("domain:%s,service:%s,data:%s\n", cs.Domain, cs.Service, string(cs.ServiceData))

	// 请参考 docs/guide/web-socket-api.md 中的定义
	// 如果消息类型持续增多，请拆分
	if cs.Service == serviceDiscover { // 写死的发现命令，优先级最高，忽略 domain，发送给所有插件
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ch := plugin.GetGlobalClient().DevicesDiscover(ctx)
		for result := range ch {
			resp := callResponse{
				ID:      cs.ID,
				Success: true,
			}
			_, err = entity.GetPluginDevice(user.AreaID, result.PluginID, result.Identity)
			if errors2.Is(err, gorm.ErrRecordNotFound) {
				resp.AddResult("device", result)
				msg, _ := json.Marshal(resp)
				cli.send <- msg
			}
		}
		return
	}

	if cs.Domain == domainPlugin { // 写死的插件命令

		var d = struct {
			PluginID string `json:"plugin_id"`
		}{}

		resp := callResponse{
			ID:      cs.ID,
			Success: true,
		}
		json.Unmarshal(cs.ServiceData, &d)
		err = nil
		var plg *plugin.Plugin
		plg, err = plugin.GetGlobalManager().Get(d.PluginID)
		if err == nil {
			switch cs.Service {
			case serviceInstall: // 插件安装
				err = plg.Install()
			case serviceUpdate:
				err = plg.Update()
			case serviceRemove:
				err = plg.Remove()
			default:
				err = errors.New(status.PluginServiceNotExist)
			}
		}
		if err != nil {
			resp.Success = false
			resp.Error = err.Error()
		}
		msg, _ := json.Marshal(resp)
		cli.send <- msg
		return err
	}

	// cmd to plugin server
	resp := callResponse{
		ID:   cs.ID,
		Type: "response",
	}
	defer func() {
		msg, _ := json.Marshal(resp)
		cli.send <- msg
		logger.Debug(string(msg))
	}()

	logger.Debug("identity:", cs.Identity)
	if cs.Service == serviceGetAttributes { // 获取设备所有属性
		d, err := plugin.GetUserDeviceAttributes(user.AreaID, user.UserID, cs.Domain, cs.Identity)
		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Success = true
			resp.AddResult("device", d)
		}
	}
	if cs.Service == serviceSetAttributes {
		// 根据插件配置判断用户是否具有权限
		if !device.IsDeviceControlPermit(user.AreaID, user.UserID, cs.Domain, cs.Identity, cs.ServiceData) {
			err = errors.New(status.Deny)
			resp.Error = err.Error()
			return
		}
		err = plugin.SetAttributes(user.AreaID, cs.Domain, cs.Identity, cs.ServiceData)
		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Success = true
		}
	}
	return err
}

// readWS
func (cli *client) readWS(user *session.User) {
	defer func() { cli.bucket.unregister <- cli }()

	for {
		t, data, err := cli.conn.ReadMessage()
		if err != nil {
			return
		}
		if t == ws.CloseMessage {
			return
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error(r)
				}
			}()
			if err := cli.handleWsMessage(data, user); err != nil {
				logger.Warnf("handle websocket message error: %s", err.Error())
			}
		}()
	}
}

// writeWS
func (cli *client) writeWS() {
	defer func() { cli.bucket.unregister <- cli }()

	for {
		select {
		case msg, ok := <-cli.send:
			if !ok {
				cli.conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}
			cli.conn.WriteMessage(ws.TextMessage, msg)
		}
	}
}

func init() {
	_callServicePool.New = func() interface{} {
		return &callService{}
	}
}
