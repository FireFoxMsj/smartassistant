package handlers

import (
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
)

func wrapRole(role orm.Role) (roleInfo, error) {

	r := roleInfo{
		RoleInfo: orm.RoleInfo{
			ID:   role.ID,
			Name: role.Name,
		},
		IsManager: role.IsManager,
	}

	// 请求数据库判断是否有改权限
	ps, err := wrapRolePermissions(role)
	if err != nil {
		return roleInfo{}, err
	}
	r.Permissions = &ps
	return r, nil
}

func wrapRolePermissions(role orm.Role) (ps Permissions, err error) {

	ps, err = getPermissions()
	if err != nil {
		return
	}
	wrapPermissions(role, ps.Device)
	for _, a := range ps.DeviceAdvanced.Locations {
		for _, d := range a.Devices {
			wrapPermissions(role, d.Permissions)
		}
	}
	wrapPermissions(role, ps.Area)
	wrapPermissions(role, ps.Location)
	wrapPermissions(role, ps.Role)
	wrapPermissions(role, ps.Scene)
	return
}

// wrapPermissions 根据权限更新配置
func wrapPermissions(role orm.Role, ps []Permission) {
	for i, v := range ps {
		ps[i].Allow = orm.IsPermit(role.ID, v.Permission.Action, v.Permission.Target, v.Permission.Attribute)
	}
}

// getPermissions 获取所有可配置的权限
func getPermissions() (Permissions, error) {

	locations, err := getLocationsWithDevice()
	if err != nil {
		return Permissions{}, err
	}
	return Permissions{
		Device:         wrapPs(permission.DevicePermission),
		DeviceAdvanced: DeviceAdvanced{Locations: locations},
		Area:           wrapPs(permission.AreaPermission),
		Location:       wrapPs(permission.LocationPermission),
		Role:           wrapPs(permission.RolePermission),
		Scene:          wrapPs(permission.ScenePermission),
	}, nil
}

func wrapPs(ps []permission.Permission) []Permission {
	var res []Permission

	for _, v := range ps {
		var a Permission
		a.Permission = v
		res = append(res, a)
	}
	return res
}

func getLocationsWithDevice() (locations []Location, err error) {

	devices, err := orm.GetDevices()
	if err != nil {
		return
	}
	// 按区域划分
	locationDevice := make(map[int][]Device)
	for _, d := range devices {
		dd := Device{Name: d.Name}
		dd.Permissions = wrapPs(orm.DevicePermissions(d))
		if _, ok := locationDevice[d.LocationID]; !ok {
			locationDevice[d.LocationID] = []Device{dd}
		} else {
			locationDevice[d.LocationID] = append(locationDevice[d.LocationID], dd)
		}
	}

	for locationID, devices := range locationDevice {

		a, _ := orm.GetLocationByID(locationID)
		aa := Location{
			Name:    a.Name,
			Devices: devices,
		}
		if aa.Name == "" {
			aa.Name = "其他"
		}
		locations = append(locations, aa)
	}
	return
}
