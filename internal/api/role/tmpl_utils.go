package role

import (
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/api/device"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types"
)

func wrapRole(role entity.Role) (roleInfo, error) {

	r := roleInfo{
		RoleInfo: entity.RoleInfo{
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

func wrapRolePermissions(role entity.Role) (ps Permissions, err error) {

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
func wrapPermissions(role entity.Role, ps []Permission) {
	for i, v := range ps {
		ps[i].Allow = entity.IsPermit(role.ID, v.Permission.Action, v.Permission.Target, v.Permission.Attribute)
	}
}

// getPermissions 获取所有可配置的权限
func getPermissions() (Permissions, error) {

	locations, err := getLocationsWithDevice()
	if err != nil {
		return Permissions{}, err
	}
	return Permissions{
		Device:         wrapPs(types.DevicePermission),
		DeviceAdvanced: DeviceAdvanced{Locations: locations},
		Area:           wrapPs(types.AreaPermission),
		Location:       wrapPs(types.LocationPermission),
		Role:           wrapPs(types.RolePermission),
		Scene:          wrapPs(types.ScenePermission),
	}, nil
}

func wrapPs(ps []types.Permission) []Permission {
	var res []Permission

	for _, v := range ps {
		var a Permission
		a.Permission = v
		res = append(res, a)
	}
	return res
}

func getLocationsWithDevice() (locations []Location, err error) {

	devices, err := entity.GetDevices()
	if err != nil {
		return
	}
	// 按区域划分
	var locationDevice sync.Map
	var wg sync.WaitGroup
	wg.Add(len(devices))
	for _, d := range devices {
		go func(d entity.Device) {
			defer wg.Done()
			ps, e := device.DevicePermissions(d)
			if e != nil {
				logrus.Error("DevicePermissionsErr:", e.Error())
				return
			}
			dd := Device{Name: d.Name, Permissions: wrapPs(ps)}
			value, ok := locationDevice.LoadOrStore(d.LocationID, []Device{dd})
			if ok {
				location := value.([]Device)
				location = append(location, dd)
			}
		}(d)
	}
	wg.Wait()
	locationDevice.Range(func(key, value interface{}) bool {
		a, _ := entity.GetLocationByID(key.(int))
		aa := Location{
			Name:    a.Name,
			Devices: value.([]Device),
		}
		if aa.Name == "" {
			aa.Name = "其他"
		}
		locations = append(locations, aa)
		return true
	})

	return
}
