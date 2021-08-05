package websocket

import (
	"encoding/json"
)

const (
	domainPlugin = "plugin"

	serviceInstall  = "install"
	serviceUpdate   = "update"
	serviceRemove   = "remove"
	serviceDiscover = "discover"

	// serviceGetAttributes 获取设备所有属性
	serviceGetAttributes = "get_attributes"
	// serviceSetAttributes 设置设备属性
	serviceSetAttributes = "set_attributes"
)

type callService struct {
	Domain      string
	ID          int
	Service     string
	ServiceData json.RawMessage `json:"service_data"`
	DeviceID    int             `json:"device_id"`
	Identity    string
	Type        string // CallService
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
	Type    string                 `json:"type"`
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
