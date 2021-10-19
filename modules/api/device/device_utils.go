package device

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	plugin2 "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

func checkDeviceName(name string) (err error) {

	if name == "" || strings.TrimSpace(name) == "" {
		err = errors.Wrap(err, status.DeviceNameInputNilErr)
		return
	}

	if utf8.RuneCountInString(name) > 20 {
		err = errors.Wrap(err, status.DeviceNameLengthLimit)
		return
	}
	return
}

// isPermit 判断用户是否有权限
func isPermit(c *gin.Context, p types.Permission) bool {
	u := session.Get(c)
	return u != nil && entity.JudgePermit(u.UserID, p)
}

// ControlPermissions 根据配置获取设备所有控制权限
func ControlPermissions(d entity.Device) ([]types.Permission, error) {
	as, err := plugin.GetControlAttributes(d)
	if err != nil {
		logger.Error("GetControlAttributesErr", err)
		return nil, err
	}
	target := types.DeviceTarget(d.ID)
	res := make([]types.Permission, 0)
	for _, attr := range as {
		name := attr.Attribute.Attribute
		attribute := entity.PluginDeviceAttr(attr.InstanceID, attr.Attribute.Attribute)
		p := types.Permission{Name: name, Action: "control", Target: target, Attribute: attribute}
		res = append(res, p)
	}
	return res, nil
}

// DevicePermissions 根据配置获取设备所有权限
func DevicePermissions(d entity.Device) (ps []types.Permission, err error) {
	ps, err = ControlPermissions(d)
	if err != nil {
		return
	}
	ps = append(ps, types.NewDeviceUpdate(d.ID))
	// 非SA设备可配置删除设备权限
	if d.Model != types.SaModel {
		ps = append(ps, types.NewDeviceDelete(d.ID))
		ps = append(ps, DeviceManagePermissions(d)...)
	}
	return
}

// IsDeviceControlPermit 控制设备的websocket命令 是否有权限
func IsDeviceControlPermit(areaID uint64, userID int, pluginID, identity string, data json.RawMessage) bool {
	d, err := entity.GetPluginDevice(areaID, pluginID, identity)
	if err != nil {
		err = errors.New(status.DeviceNotExist)
		logger.Warning(err)
		return false
	}

	var req plugin2.SetRequest
	json.Unmarshal(data, &req)
	for _, attr := range req.Attributes {
		logger.Debug(d, attr)
		if !entity.IsDeviceControlPermitByAttr(userID, d.ID, attr.InstanceID, attr.Attribute) {
			return false
		}
	}
	return true
}

// DeviceManagePermissions 设备的管理权限，暂时只有固件升级
func DeviceManagePermissions(d entity.Device) []types.Permission {
	var permissions = make([]types.Permission, 0)
	permissions = append(permissions, types.NewDeviceManage(d.ID))
	return permissions
}
