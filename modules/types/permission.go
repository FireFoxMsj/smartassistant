package types

import (
	"fmt"
)

const (
	FwUpgrade       = "firmware_upgrade" // 固件升级
	SoftwareUpgrade = "software_upgrade" // 软件升级
)

type Permission struct {
	Name      string `json:"name"`
	Action    string `json:"action"`    // 动作
	Target    string `json:"target"`    // 对象
	Attribute string `json:"attribute"` // 属性
}

var (
	// 设备
	DeviceAdd     = Permission{"添加设备", "add", "device", ""}
	DeviceUpdate  = Permission{"修改设备", "update", "device", ""}
	DeviceControl = Permission{"控制设备", "control", "device", ""}
	DeviceDelete  = Permission{"删除设备", "delete", "device", ""}
	DeviceManage  = Permission{"管理设备", "manage", "device", ""}
	// 家庭/公司
	AreaGetCode          = Permission{"生成邀请码", "get", "area", "invite_code"}
	AreaUpdateName       = Permission{"修改名称", "update", "area", "name"}
	AreaUpdateMemberRole = Permission{"修改成员角色", "update", "area", "member_role"}
	AreaDelMember        = Permission{"删除成员", "delete", "area", "member"}
	// 房间/区域
	LocationAdd         = Permission{"添加房间/区域", "add", "location", ""}
	LocationUpdateOrder = Permission{"调整顺序", "update", "location", "order"}
	LocationUpdateName  = Permission{"修改房间名称", "update", "location", "name"}
	LocationGet         = Permission{"查看房间详情", "get", "location", ""}
	LocationDel         = Permission{"删除房间", "delete", "location", ""}
	// 角色
	RoleGet    = Permission{"查看角色列表", "get", "role", ""}
	RoleAdd    = Permission{"新增角色", "add", "role", ""}
	RoleUpdate = Permission{"编辑角色", "update", "role", ""}
	RoleDel    = Permission{"删除角色", "delete", "role", ""}

	// 场景
	SceneAdd     = Permission{"新增场景", "add", "scene", ""}
	SceneUpdate  = Permission{"修改场景", "update", "scene", ""}
	SceneDel     = Permission{"删除场景", "delete", "scene", ""}
	SceneControl = Permission{"控制场景", "control", "scene", ""}

	DevicePermission   = []Permission{DeviceAdd, DeviceUpdate, DeviceControl, DeviceDelete, DeviceManage}
	AreaPermission     = []Permission{AreaGetCode, AreaUpdateName, AreaUpdateMemberRole, AreaDelMember}
	LocationPermission = []Permission{LocationAdd, LocationUpdateOrder, LocationUpdateName, LocationGet, LocationDel}
	RolePermission     = []Permission{RoleGet, RoleAdd, RoleUpdate, RoleDel}
	ScenePermission    = []Permission{SceneAdd, SceneUpdate, SceneDel, SceneControl}

	DefaultPermission []Permission

	ManagerPermission []Permission
	MemberPermission  []Permission
)

func DeviceTarget(deviceID int) string {
	return fmt.Sprintf("device-%d", deviceID)
}

func NewDeviceDelete(deviceID int) Permission {
	target := DeviceTarget(deviceID)
	return Permission{"删除设备", "delete", target, ""}
}
func NewDeviceUpdate(deviceID int) Permission {
	target := DeviceTarget(deviceID)
	return Permission{"修改设备", "update", target, ""}
}

func NewDeviceManage(deviceID int, name string, attr string) Permission {
	target := DeviceTarget(deviceID)
	return Permission{name, "manage", target, attr}
}

func init() {

	DefaultPermission = append(DefaultPermission, DevicePermission...)
	DefaultPermission = append(DefaultPermission, AreaPermission...)
	DefaultPermission = append(DefaultPermission, LocationPermission...)
	DefaultPermission = append(DefaultPermission, RolePermission...)
	DefaultPermission = append(DefaultPermission, ScenePermission...)

	ManagerPermission = append(ManagerPermission, DefaultPermission...)
	MemberPermission = []Permission{DeviceControl, LocationGet}
}
