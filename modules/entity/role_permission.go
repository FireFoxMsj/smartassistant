package entity

import (
	"fmt"

	"github.com/zhiting-tech/smartassistant/modules/types"
)

type ActionType string

// RolePermission
type RolePermission struct {
	ID        int
	RoleID    int  `gorm:"index:permission,unique"` // 角色
	Role      Role `gorm:"constraint:OnDelete:CASCADE;"`
	Name      string
	Action    string `gorm:"index:permission,unique"` // 动作
	Target    string `gorm:"index:permission,unique"` // 对象
	Attribute string `gorm:"index:permission,unique"` // 属性
}

func (p RolePermission) TableName() string {
	return "role_permissions"
}

// IsDeviceControlPermit 判断用户是否有该设备的某个控制权限
func IsDeviceControlPermit(userID, deviceID int, attr Attribute) bool {
	return IsDeviceControlPermitByAttr(userID, deviceID, attr.InstanceID, attr.Attribute.Attribute)
}

func PluginDeviceAttr(instanceID int, attr string) string {
	return fmt.Sprintf("%d_%s", instanceID, attr)
}

// IsDeviceControlPermitByAttr 判断用户是否有该设备的某个控制权限
func IsDeviceControlPermitByAttr(userID, deviceID, instanceID int, attr string) bool {
	target := types.DeviceTarget(deviceID)
	return judgePermit(userID, "control", target, PluginDeviceAttr(instanceID, attr))
}

type Attr struct {
	DeviceID   int
	InstanceID int
	Attribute  string
}

type UserPermissions struct {
	ps []RolePermission
}

func (up UserPermissions) IsDeviceControlPermit(deviceID, instanceID int, attr string) bool {
	for _, p := range up.ps {
		if p.Action == "control" &&
			p.Target == types.DeviceTarget(deviceID) &&
			p.Attribute == PluginDeviceAttr(instanceID, attr) {
			return true
		}
	}
	return false
}

func (up UserPermissions) IsDeviceAttrPermit(deviceID int, attr Attribute) bool {
	for _, p := range up.ps {
		if p.Action == "control" &&
			p.Target == types.DeviceTarget(deviceID) &&
			p.Attribute == PluginDeviceAttr(attr.InstanceID, attr.Attribute.Attribute) {
			return true
		}
	}
	return false
}

// GetUserPermissions 获取用户的所有权限
func GetUserPermissions(userID int) (up UserPermissions, err error) {

	roleIds, err := GetRoleIdsByUid(userID)
	if err != nil {
		return
	}
	var ps []RolePermission
	if err = GetDB().Joins("Role", db.Where("id in ?", roleIds)).
		Find(&ps).Error; err != nil {
		return
	}
	return UserPermissions{ps: ps}, nil
}

func JudgePermit(userID int, p types.Permission) bool {
	return judgePermit(userID, p.Action, p.Target, p.Attribute)
}

// DeviceControlPermit 判断用户是否有设备的任一控制权限
func DeviceControlPermit(userID, deviceID int) bool {
	roleIds, err := GetRoleIdsByUid(userID)
	if err != nil {
		return false
	}

	if len(roleIds) == 0 {
		return false
	}

	var permissions []RolePermission

	if err := GetDB().Where("role_id in ? and action = ? and target = ?",
		roleIds, "control", types.DeviceTarget(deviceID)).Find(&permissions).Error; err != nil {
		return false
	}
	if len(permissions) == 0 {
		return false
	}

	return true
}
func judgePermit(userID int, action, target, attribute string) bool {
	// SA拥有者默认拥有所有权限
	if IsAreaOwner(userID) {
		return true
	}
	roleIds, err := GetRoleIdsByUid(userID)
	if err != nil {
		return false
	}

	if len(roleIds) == 0 {
		return false
	}

	var permissions []RolePermission

	if err := GetDB().Where("role_id in ? and action = ? and target = ? and attribute = ?",
		roleIds, action, target, attribute).Find(&permissions).Error; err != nil {
		return false
	}

	if len(permissions) == 0 {
		return false
	}

	return true
}

func IsPermit(roleID int, action, target, attribute string) bool {
	p := RolePermission{
		RoleID:    roleID,
		Action:    action,
		Target:    target,
		Attribute: attribute,
	}
	if err := GetDB().First(&p, p).Error; err != nil {
		return false
	}
	return true
}

func IsDeviceActionPermit(roleID int, action string) bool {
	return IsPermit(roleID, action, "device", "")
}
