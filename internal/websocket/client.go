package websocket

import (
	"context"
	"encoding/json"
	errors2 "errors"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/api/device"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/plugin"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

type client struct {
	key    string
	conn   *ws.Conn
	send   chan []byte
	bucket *bucket
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

	logrus.Printf("cs:%s ,%s ,%s\n", cs.Domain, cs.Service, string(cs.ServiceData))
	plgManager := plugin.GetManager()

	// 请参考 docs/guide/web-socket-api.md 中的定义
	// 如果消息类型持续增多，请拆分
	if cs.Service == serviceDiscover { // 写死的发现命令，优先级最高，忽略 domain，发送给所有插件
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ch := plgManager.DeviceDiscover(ctx)
		for result := range ch {
			resp := callResponse{
				ID:      cs.ID,
				Success: true,
			}
			_, err = entity.GetDeviceByIdentity(result.Identity)
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
		json.Unmarshal(cs.ServiceData, &d)
		err = nil
		switch cs.Service {
		case serviceInstall: // 插件安装
			err = plgManager.PluginInstall(d.PluginID)
		case serviceUpdate:
			err = plgManager.PluginUpdate(d.PluginID)
		case serviceRemove:
			err = plgManager.PluginRemove(d.PluginID)
		default:
			err = errors.New(status.PluginServiceNotExist)
		}

		resp := callResponse{
			ID:      cs.ID,
			Success: true,
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
		logrus.Debug(string(msg))
	}()

	logrus.Debug("identity:", cs.Identity)
	if cs.Service == serviceGetAttributes { // 获取设备所有属性
		d, err := plugin.GetUserDeviceAttributes(user.UserID, cs.Domain, cs.Identity)
		if err != nil {
			resp.Error = err.Error()
		} else {
			resp.Success = true
			resp.AddResult("device", d)
		}
	}
	if cs.Service == serviceSetAttributes {
		// 根据插件配置判断用户是否具有权限
		if !device.IsDeviceControlPermit(user.UserID, cs.Identity, cs.ServiceData) {
			err = errors.New(status.Deny)
			resp.Error = err.Error()
			return
		}
		err = plugin.SetAttributes(cs.Domain, cs.Identity, cs.ServiceData)
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
					logrus.Error(r)
				}
			}()
			if err := cli.handleWsMessage(data, user); err != nil {
				logrus.Warnf("handle websocket message error: %s", err.Error())
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
