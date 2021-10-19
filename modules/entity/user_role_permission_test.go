package entity

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils"

	"strconv"
	"testing"
)

func TestAddManagerRole(t *testing.T) {
	ast := assert.New(t)

	const managerRoleName = "npc"
	const lengthGreaterThan20RoleName = "123456789123456789123456789"

	areaID := utils.SAAreaID()
	role, err := AddManagerRoleWithDB(GetDB(), managerRoleName, areaID)
	ast.NoError(err, "add manager role with db error: %v", err)
	ast.NotEmpty(role)

	role, err = AddManagerRoleWithDB(GetDB(), managerRoleName, areaID)
	ast.NoError(err, "add manager role with db error: %v", err)
	ast.NotEmpty(role)

	_, err = AddManagerRoleWithDB(GetDB(), lengthGreaterThan20RoleName, areaID)
	ast.Error(err, "add manager role with db error: %v", err)
}

func TestAddRole(t *testing.T) {
	ast := assert.New(t)

	const roleName = "cc"
	areaID := utils.SAAreaID()
	role, err := AddRole(roleName, areaID)
	ast.NoError(err, "add role error: %v", err)
	ast.NotEmpty(role)

	role, err = AddRole(roleName, areaID)
	ast.NoError(err, "add role error: %v", err)
	ast.NotEmpty(role)
}

func TestGetRole(t *testing.T) {
	ast := assert.New(t)

	const existRoleID = 1
	const notExistRoleID = 9999

	role, err := GetRoleByID(existRoleID)
	ast.NoError(err, "get role by id error: %v", err)
	ast.NotEmpty(role)

	role, err = GetRoleByID(notExistRoleID)
	ast.Error(err, "get role by id error")
	ast.Empty(role)

	// 填入areaID
	areaID := utils.SAAreaID()
	roles, err := GetRoles(areaID)
	ast.NoError(err, "get roles error: %v", err)
	ast.NotEmpty(roles)

	roles, err = GetRolesByIds([]int{})
	ast.NoError(err, "get roles error: %v", err)
	ast.Empty(roles)

	roles, err = GetRolesByIds([]int{1, 2})
	ast.NoError(err, "get roles by ids error: %v", err)
	ast.NotEmpty(roles)

	roles, err = GetRolesByIds([]int{9999, 8888})
	ast.NoError(err, "get roles error: %v", err)
	ast.Empty(roles)

	roles, err = GetRolesByIds([]int{1, 8888})
	ast.NoError(err, "get roles by ids error: %v", err)
	ast.NotEmpty(roles)
}

func TestIsRoleNameExist(t *testing.T) {
	ast := assert.New(t)

	var ts = []struct {
		name        string
		roleID      int
		expectedRes bool
	}{
		{"npc", 0, true},
		{"npc", 1, false},
		{"npc", 2, true},
		{"npc", 999, true},
		{"cc", 0, true},
		{"", 0, false},
		{"cc", 1, true},
		{"cc", 2, false},
	}

	areaID := utils.SAAreaID()
	for _, t := range ts {
		ast.Equal(t.expectedRes, IsRoleNameExist(t.name, t.roleID, areaID))
	}
}

func TestUpdateRole(t *testing.T) {
	ast := assert.New(t)

	const existManagerRoleID = 1
	const existRoleID = 2
	const notExistRoleID = 9999
	const newRoleName = "new"

	_, err := UpdateRole(existManagerRoleID, newRoleName)
	ast.Error(err, "update role error: %v")

	role, err := UpdateRole(existRoleID, newRoleName)
	ast.NoError(err, "update role error: %v")
	ast.Equal(newRoleName, role.Name)

	_, err = UpdateRole(notExistRoleID, newRoleName)
	ast.Error(err, "update role error")
}

func TestRole_AddPermissionForRole(t *testing.T) {
	ast := assert.New(t)

	var tt = []struct {
		name      string
		action    string
		target    string
		attribute string
	}{
		{"控制开关", "control", "device-2", "power"},
		{"控制调节亮度", "control", "device-2", "brightness"},
		{"修改设备", "update", "device-2", ""},
		{"删除设备", "delete", "device-2", ""},
		{"控制开关", "control", "device-2", "power"},
	}

	managerRole, _ := GetRoleByID(1)
	role, _ := GetRoleByID(2)

	for _, t := range tt {
		err := managerRole.AddPermissionForRole(t.name, t.action, t.target, t.attribute)
		ast.NoError(err, "add permission for role: %v", err)
	}

	for _, t := range tt {
		err := role.AddPermissionForRole(t.name, t.action, t.target, t.attribute)
		ast.NoError(err, "add permission for role: %v", err)
	}
}

func TestRole_AddPermission(t *testing.T) {
	ast := assert.New(t)

	managerRole, _ := GetRoleByID(1)
	role, _ := GetRoleByID(2)

	err := managerRole.addPermission(GetDB(), types.DeviceAdd)
	ast.NoError(err, "add permission error: %v", err)
	err = managerRole.addPermission(GetDB(), types.DeviceAdd)
	ast.NoError(err, "add permission error: %v", err)

	err = role.addPermission(GetDB(), types.DeviceAdd)
	ast.NoError(err, "add permission error: %v", err)
	err = role.addPermission(GetDB(), types.DeviceAdd)
	ast.NoError(err, "add permission error: %v", err)
}

func TestInitRole(t *testing.T) {
	ast := assert.New(t)
	areaID := utils.SAAreaID()
	err := InitRole(GetDB(), areaID)
	ast.NoError(err, "init role error: %v", err)
}

func TestCreateUserRole(t *testing.T) {
	ast := assert.New(t)

	var urs = []UserRole{
		{UserID: 1, RoleID: 1},
		{UserID: 1, RoleID: 2},
		{UserID: 1, RoleID: 3},
		{UserID: 2, RoleID: 4},
		{UserID: 3, RoleID: 4},
	}

	var emptyUrs = []UserRole{}

	err := CreateUserRole(urs)
	ast.NoError(err, "create UserRole error: %v", err)

	err = CreateUserRole(emptyUrs)
	ast.Error(err, "create UserRole error: %v")
}

func TestGetRoleIdsByUid(t *testing.T) {
	ast := assert.New(t)

	const existUID1 = 1
	const existUID2 = 2
	const notExistUID = 999

	roleIds, err := GetRoleIdsByUid(existUID1)
	ast.NoError(err, "get role ids by uid error: %v", err)
	ast.GreaterOrEqual(len(roleIds), 1)

	roleIds, err = GetRoleIdsByUid(existUID2)
	ast.NoError(err, "get role ids by uid error: %v", err)
	ast.Equal(len(roleIds), 1)

	roleIds, err = GetRoleIdsByUid(notExistUID)
	ast.NoError(err, "get role ids by uid error: %v", err)
	ast.Empty(roleIds)
}

func TestGetRolesByUid(t *testing.T) {
	ast := assert.New(t)

	const existUID1 = 1
	const existUID2 = 2
	const notExistUID = 999
	areaID := utils.SAAreaID()
	for i := 0; i < 3; i++ {
		name := strconv.Itoa(i)
		_, _ = AddRole(name, areaID)
	}

	roles, err := GetRolesByUid(existUID1)
	ast.NoError(err, "get roles by uid error: %v", err)
	ast.GreaterOrEqual(len(roles), 1)

	roles, err = GetRolesByUid(existUID2)
	ast.NoError(err, "get roles by uid error: %v", err)
	ast.Equal(len(roles), 1)

	roles, err = GetRolesByUid(notExistUID)
	ast.NoError(err, "get roles by uid error: %v", err)
	ast.Empty(roles)
}

func TestIsPermit(t *testing.T) {
	ast := assert.New(t)

	const managerID = 3
	const memberID = 4
	const notExistID = 999

	var tt = []struct {
		roleID      int
		permission  types.Permission
		expectedRes bool
	}{
		{managerID, types.DeviceAdd, true},
		{managerID, types.RoleAdd, true},
		{managerID, types.DeviceDelete, true},
		{memberID, types.DeviceAdd, false},
		{memberID, types.DeviceControl, true},
		{memberID, types.LocationGet, true},
		{notExistID, types.DeviceAdd, false},
	}

	for _, t := range tt {
		res := IsPermit(t.roleID, t.permission.Action, t.permission.Target, t.permission.Attribute)
		ast.Equal(res, t.expectedRes)
	}
}

func TestIsDeviceControlPermit(t *testing.T) {
	ast := assert.New(t)

	const managerID = 3
	const memberID = 4
	const notExistID = 999

	var tt = []struct {
		roleID      int
		permission  types.Permission
		expectedRes bool
	}{
		{managerID, types.DeviceAdd, true},
		{managerID, types.DeviceDelete, true},
		{memberID, types.DeviceAdd, false},
		{memberID, types.DeviceControl, true},
		{notExistID, types.DeviceAdd, false},
	}

	for _, t := range tt {
		res := IsDeviceActionPermit(t.roleID, t.permission.Action)
		ast.Equal(res, t.expectedRes)
	}
}

func TestJudgePermit(t *testing.T) {
	ast := assert.New(t)

	const managerUserID = 1
	const memberUserID = 2
	const notExistUserID = 999

	var tt = []struct {
		userID      int
		permission  types.Permission
		expectedRes bool
	}{
		{managerUserID, types.DeviceAdd, true},
		{managerUserID, types.RoleAdd, true},
		{managerUserID, types.DeviceDelete, true},
		{memberUserID, types.DeviceAdd, false},
		{memberUserID, types.DeviceControl, true},
		{memberUserID, types.LocationGet, true},
		{notExistUserID, types.DeviceAdd, false},
	}

	for _, t := range tt {
		res := judgePermit(t.userID, t.permission.Action, t.permission.Target, t.permission.Attribute)
		ast.Equal(t.expectedRes, res)

		res = JudgePermit(t.userID, t.permission)
		ast.Equal(t.expectedRes, res)
	}
}

func TestDeviceControlPermit(t *testing.T) {
	ast := assert.New(t)

	const managerUserID = 1
	const memberUserID = 2
	const notExistUserID = 999
	const exitDeviceID = 2
	const notExitDeviceID = 999

	var ss = []struct {
		name      string
		action    string
		target    string
		attribute string
	}{
		{"控制开关", "control", "device-2", "power"},
		{"控制调节亮度", "control", "device-2", "brightness"},
		{"修改设备", "update", "device-2", ""},
		{"删除设备", "delete", "device-2", ""},
	}

	managerRole, _ := GetRoleByID(3)

	for _, s := range ss {
		_ = managerRole.AddPermissionForRole(s.name, s.action, s.target, s.attribute)
	}

	var tt = []struct {
		userID      int
		deviceID    int
		expectedRes bool
	}{
		{managerUserID, exitDeviceID, true},
		{managerUserID, notExitDeviceID, false},
		{memberUserID, exitDeviceID, false},
		{memberUserID, notExitDeviceID, false},
		{notExistUserID, notExitDeviceID, false},
	}

	for _, t := range tt {
		res := DeviceControlPermit(t.userID, t.deviceID)
		ast.Equal(t.expectedRes, res)
	}
}

func TestRole_DelPermission(t *testing.T) {
	ast := assert.New(t)

	managerRole, _ := GetRoleByID(1)
	role, _ := GetRoleByID(2)

	err := managerRole.DelPermission(types.DeviceAdd)
	ast.NoError(err, "delete permission error: %v", err)

	err = managerRole.DelPermission(types.RoleAdd)
	ast.NoError(err, "delete permission error: %v", err)

	err = role.DelPermission(types.DeviceAdd)
	ast.NoError(err, "delete permission error: %v", err)

	err = role.DelPermission(types.RoleAdd)
	ast.NoError(err, "delete permission error: %v", err)
}

func TestDelUserRoleByUid(t *testing.T) {
	ast := assert.New(t)

	const existUID1 = 1
	const existUID2 = 2
	const notExistUID = 999

	err := DelUserRoleByUid(existUID1, GetDB())
	ast.NoError(err, "delete UserRole by Uid error: %v", err)
	roleIDs, _ := GetRoleIdsByUid(existUID1)
	ast.Equal(len(roleIDs), 0)

	err = DelUserRoleByUid(existUID2, GetDB())
	ast.NoError(err, "delete UserRole by Uid error: %v", err)
	roleIDs, _ = GetRoleIdsByUid(existUID1)
	ast.Equal(len(roleIDs), 0)

	err = DelUserRoleByUid(notExistUID, GetDB())
	ast.NoError(err, "delete UserRole by Uid error: %v", err)
	roleIDs, _ = GetRoleIdsByUid(existUID1)
	ast.Equal(len(roleIDs), 0)
}

func TestDeleteRole(t *testing.T) {
	ast := assert.New(t)

	const exitManagerRID = 1
	const exitMemberRID = 2
	const notExistRID = 999

	err := DeleteRole(exitManagerRID)
	ast.Error(err, "delete role error")
	role, _ := GetRoleByID(exitManagerRID)
	ast.NotEmpty(role)

	err = DeleteRole(exitMemberRID)
	ast.NoError(err, "delete role error: %v", err)
	role, _ = GetRoleByID(exitMemberRID)
	ast.Empty(role)

	err = DeleteRole(notExistRID)
	ast.Error(err, "delete role error: %v", err)
}
