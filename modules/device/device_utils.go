package device

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	plugin2 "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

// IsPermit 判断用户是否有权限
func IsPermit(c *gin.Context, p types.Permission) bool {
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

// Permissions 根据配置获取设备所有权限
func Permissions(d entity.Device) (ps []types.Permission, err error) {
	ps = append(ps, ManagePermissions(d)...)
	ps = append(ps, types.NewDeviceUpdate(d.ID))

	if d.Model == types.SaModel {
		return
	}

	controlPermission, err := ControlPermissions(d)
	if err != nil {
		return
	}

	// 非SA设备可配置删除设备权限,控制设备权限
	ps = append(ps, controlPermission...)
	ps = append(ps, types.NewDeviceDelete(d.ID))
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
	if err = json.Unmarshal(data, &req); err != nil {
		logrus.Errorf("IsDeviceControlPermit unmarshal err: %s", err.Error())
		return false
	}
	up, err := entity.GetUserPermissions(userID)
	if err != nil {
		return false
	}
	for _, attr := range req.Attributes {
		logger.Debug(d, attr)
		if !up.IsDeviceAttrControlPermit(d.ID, attr.InstanceID, attr.Attribute) {
			return false
		}
	}
	return true
}

// ManagePermissions 设备的管理权限
func ManagePermissions(d entity.Device) []types.Permission {
	var permissions = make([]types.Permission, 0)
	// TODO 设备的固件升级功能是否能和设备的其他控制属性一样从插件获取？
	if d.Model == types.SaModel {
		permissions = append(permissions, types.NewDeviceManage(d.ID, "固件升级", types.FwUpgrade))
		permissions = append(permissions, types.NewDeviceManage(d.ID, "软件升级", types.SoftwareUpgrade))
	}
	return permissions
}
