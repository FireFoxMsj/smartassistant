package websocket

import (
	"encoding/json"
	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

func GetAttrs(cs callService) (result Result, err error) {
	result = make(Result)
	user := cs.CallUser
	d, err := plugin.GetUserDeviceAttributes(user.AreaID, user.UserID, cs.Domain, cs.Identity)
	if err != nil {
		return
	}
	result["device"] = d
	return
}
func SetAttrs(cs callService) (result Result, err error) {

	result = make(Result)
	user := cs.CallUser
	_, err = entity.GetDeviceByIdentity(cs.Identity)
	if err == nil {
		// 根据插件配置判断用户是否具有权限
		if !device.IsDeviceControlPermit(user.AreaID, user.UserID, cs.Domain, cs.Identity, cs.ServiceData) {
			err = errors.New(status.Deny)
			return
		}
	} else {
		if err != gorm.ErrRecordNotFound {
			return
		}
		err = nil
	}

	err = plugin.SetAttributes(user.AreaID, cs.Domain, cs.Identity, cs.ServiceData)
	if err != nil {
		return
	}
	return
}

// ConnectDevice 连接设备 TODO 直接替代添加设备接口？
func ConnectDevice(cs callService) (result Result, err error) {
	result = make(Result)
	var authParams map[string]string
	if err = json.Unmarshal(cs.ServiceData, &authParams); err != nil {
		return
	}
	d, err := plugin.ConnectDevice(cs.Identity, cs.Domain, authParams)
	if err != nil {
		return
	}
	result["device"] = d

	// 自动加入设备列表
	deviceEntity, err := plugin.GetInfoFromDeviceAttrs(cs.Domain, d)
	if err != nil {
		return
	}
	if err = device.Create(cs.CallUser.AreaID, &deviceEntity); err != nil {
		return
	}
	return
}

// DisconnectDevice 设备断开连接（取消配对等） TODO 直接替代删除设备接口？
func DisconnectDevice(cs callService) (result Result, err error) {

	result = make(Result)
	var authParams map[string]string
	if err = json.Unmarshal(cs.ServiceData, &authParams); err != nil {
		return
	}
	err = plugin.DisconnectDevice(cs.Identity, cs.Domain, authParams)
	if err != nil {
		return
	}
	return
}

func RegisterCmd() {
	RegisterCallFunc(serviceConnect, ConnectDevice)
	RegisterCallFunc(serviceDisconnect, DisconnectDevice)
	RegisterCallFunc(serviceSetAttributes, SetAttrs)
	RegisterCallFunc(serviceGetAttributes, GetAttrs)
}
