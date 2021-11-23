package websocket

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
)

type MsgType string

const (
	// serviceDiscover 发现设备
	serviceDiscover = "discover"

	// serviceGetAttributes 获取设备所有属性
	serviceGetAttributes = "get_attributes"
	// serviceSetAttributes 设置设备属性
	serviceSetAttributes = "set_attributes"
	// serviceConnect 连接（认证、配对）
	serviceConnect = "connect"
	// serviceDisconnect 断开连接（取消配对）
	serviceDisconnect = "disconnect"

	MsgTypeResponse MsgType = "response"
)

type callService struct {
	Domain      string
	ID          int
	Service     string
	ServiceData json.RawMessage `json:"service_data"`
	DeviceID    int             `json:"device_id"`
	Identity    string
	Type        string // CallService

	CallUser session.User
}

func (cs *callService) reset() {
	cs.Domain = ""
	cs.ID = 0
	cs.Service = ""
	cs.Type = ""
	cs.Identity = ""
}

type callResponse struct {
	ID      int                    `json:"id"`
	Type    MsgType                `json:"type"`
	Error   string                 `json:"error,omitempty"`
	Result  map[string]interface{} `json:"result"`
	Success bool                   `json:"success"`
}

func (cr *callResponse) AddResult(key string, value interface{}) {
	if cr.Result == nil {
		cr.Result = make(map[string]interface{})
	}
	cr.Result[key] = value
}

type Event struct {
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
}

type Result map[string]interface{}

type CallFunc func(service callService) (Result, error)

var callFunctions = make(map[string]CallFunc)

func RegisterCallFunc(cmd string, callFunc CallFunc) {
	if _, ok := callFunctions[cmd]; ok {
		logrus.Panic("call cmd already exist")
	}
	callFunctions[cmd] = callFunc
}
