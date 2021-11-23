package device

import (
	"encoding/json"
	"errors"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gorm.io/gorm"
)

func Create(areaID uint64, device *entity.Device) (err error) {
	if device == nil {
		return errors.New("nil device")
	}

	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {

		device.AreaID = areaID
		// Create 添加设备
		switch device.Model {
		case types.SaModel:
			// 添加设备为SA时不需要添加设备影子
			if err = entity.AddSADevice(device, tx); err != nil {
				return err
			}
		default:
			if err = saveWithThingModel(device, tx); err != nil {
				return err
			}
		}
		// 为所有角色增加改设备的权限
		return AddDevicePermissionForRoles(*device, tx)
	}); err != nil {
		return
	}
	return
}

// AddDevicePermissionForRoles 为所有角色增加设备权限
func AddDevicePermissionForRoles(device entity.Device, tx *gorm.DB) (err error) {

	// 将权限赋给给所有角色
	var roles []entity.Role
	// 使用同一个DB，保证在一个事务内
	roles, err = entity.GetRolesWithTx(tx, device.AreaID)
	if err != nil {
		return err
	}
	for _, role := range roles {
		// 查看角色设备权限模板配置
		if entity.IsDeviceActionPermit(role.ID, "manage", tx) {
			role.AddPermissionsWithDB(tx, ManagePermissions(device)...)
		}

		if entity.IsDeviceActionPermit(role.ID, "update", tx) {
			role.AddPermissionsWithDB(tx, types.NewDeviceUpdate(device.ID))
		}

		// SA设备不需要配置控制和删除权限
		if device.Model == types.SaModel {
			continue
		}
		if entity.IsDeviceActionPermit(role.ID, "control", tx) {
			var ps []types.Permission
			ps, err = ControlPermissions(device)
			if err != nil {
				logger.Error("ControlPermissionsErr:", err.Error())
				continue
			}
			role.AddPermissionsWithDB(tx, ps...)
		}

		if entity.IsDeviceActionPermit(role.ID, "delete", tx) {
			role.AddPermissionsWithDB(tx, types.NewDeviceDelete(device.ID))
		}

	}
	return
}

// saveWithThingModel 添加设备,并保存物模型和生成设备影子
func saveWithThingModel(d *entity.Device, tx *gorm.DB) (err error) {
	// 获取所有属性
	das, err := plugin.GetGlobalClient().GetAttributes(*d)
	if err != nil {
		return
	}
	// 保存物模型
	d.ThingModel, err = json.Marshal(das)
	if err != nil {
		return
	}

	// 新建设备影子并更新
	shadow := entity.NewShadow()
	for _, ins := range das.Instances {
		for _, attr := range ins.Attributes {
			shadow.UpdateReported(ins.InstanceId, attr.Attribute)
		}
	}
	d.Shadow, err = json.Marshal(shadow)
	if err != nil {
		return
	}
	if err = entity.AddDevice(d, tx); err != nil {
		return
	}
	go plugin.GetGlobalClient().HealthCheck(*d)
	return nil
}
