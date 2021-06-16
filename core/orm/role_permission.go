package orm

import (
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
)

type ActionType string

// RolePermission
type RolePermission struct {
	ID        int
	RoleID    int `gorm:"index:permission,unique"` // 角色
	Name      string
	Action    string `gorm:"index:permission,unique"` // 动作
	Target    string `gorm:"index:permission,unique"` // 对象
	Attribute string `gorm:"index:permission,unique"` // 属性
}

func (p RolePermission) TableName() string {
	return "role_permissions"
}

// IsDeviceControlPermit 判断用户是否有该设备的某个控制权限
func IsDeviceControlPermit(userID, deviceID int, attribute string) bool {
	target := permission.DeviceTarget(deviceID)
	return judgePermit(userID, "control", target, attribute)
}

func JudgePermit(userID int, p permission.Permission) bool {
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
		roleIds, "control", permission.DeviceTarget(deviceID)).Find(&permissions).Error; err != nil {
		return false
	}

	if len(permissions) == 0 {
		return false
	}

	return true
}
func judgePermit(userID int, action, target, attribute string) bool {
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
