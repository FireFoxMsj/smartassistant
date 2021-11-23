package entity

import (
	errors2 "errors"
	"gorm.io/gorm/clause"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// Device 识别的设备
type Device struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`                                                // 地址
	Identity     string    `json:"identity" gorm:"uniqueIndex:area_id_identity_plugin_id"` // 设备唯一值
	Model        string    `json:"model"`                                                  // 型号
	Manufacturer string    `json:"manufacturer"`                                           // 制造商
	Type         string    `json:"type"`                                                   // 设备类型，如：light,switch...
	PluginID     string    `json:"plugin_id" gorm:"uniqueIndex:area_id_identity_plugin_id"`
	CreatedAt    time.Time `json:"created_at"`
	LocationID   int       `json:"location_id"`
	Deleted      gorm.DeletedAt

	AreaID uint64 `json:"area_id" gorm:"type:bigint;uniqueIndex:area_id_identity_plugin_id"`
	Area   Area   `gorm:"constraint:OnDelete:CASCADE;"`

	Shadow     datatypes.JSON `json:"-"`
	ThingModel datatypes.JSON `json:"-"`
}

func (d Device) TableName() string {
	return "devices"
}

func (d *Device) AfterDelete(tx *gorm.DB) (err error) {
	// 删除设备所有相关权限
	target := types.DeviceTarget(d.ID)
	return tx.Delete(&RolePermission{}, "target = ?", target).Error
}

func GetDeviceByID(id int) (device Device, err error) {
	err = GetDB().First(&device, "id = ?", id).Error
	return
}

func GetDevicesByPluginID(pluginID string) (devices []Device, err error) {
	err = GetDB().Where(Device{PluginID: pluginID}).Find(&devices).Error
	return
}

// GetDeviceByIDWithUnscoped 获取设备，包括已删除
func GetDeviceByIDWithUnscoped(id int) (device Device, err error) {
	err = GetDB().Unscoped().First(&device, "id = ?", id).Error
	return
}

// GetPluginDevice 获取插件的设备
func GetPluginDevice(areaID uint64, pluginID, identity string) (device Device, err error) {
	filter := Device{
		Identity: identity,
		PluginID: pluginID,
	}
	err = GetDBWithAreaScope(areaID).Where(filter).First(&device).Error
	return
}

// GetManufacturerDevice 获取厂商的设备
func GetManufacturerDevice(areaID uint64, manufacturer, identity string) (device Device, err error) {
	filter := Device{
		Identity:     identity,
		Manufacturer: manufacturer,
	}
	err = GetDBWithAreaScope(areaID).Where(filter).First(&device).Error
	return
}

func GetDevices(areaID uint64) (devices []Device, err error) {
	err = GetDBWithAreaScope(areaID).Find(&devices).Error
	return
}

// GetZhitingDevices 获取所有智汀设备
func GetZhitingDevices() (devices []Device, err error) {
	err = GetDB().Where(Device{Manufacturer: "zhiting"}).Find(&devices).Error
	return
}

func GetDevicesByLocationID(locationId int) (devices []Device, err error) {
	err = GetDB().Order("created_at asc").Find(&devices, "location_id = ?", locationId).Error
	return
}

func DelDeviceByID(id int) (err error) {
	d := Device{ID: id}
	err = GetDB().Delete(&d).Error
	return
}

func DelDevicesByPlgID(plgID string) (err error) {
	err = GetDB().Delete(&Device{}, "plugin_id = ?", plgID).Error
	return
}

func UpdateDevice(id int, updateDevice Device) (err error) {
	device := &Device{ID: id}
	err = GetDB().First(device).Updates(updateDevice).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.DeviceNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func GetSaDevice() (device Device, err error) {
	err = GetDB().First(&device, "model = ?", types.SaModel).Error
	return
}

func UnBindLocationDevices(locationID int) (err error) {
	err = GetDB().Find(&Device{}, "location_id = ?", locationID).Update("location_id", 0).Error
	return
}

func UnBindLocationDevice(deviceID int) (err error) {
	device := &Device{ID: deviceID}
	err = GetDB().First(device).Update("location_id", 0).Error
	return
}

// CheckDeviceExist 设备是否已存在
func CheckDeviceExist(device Device, tx *gorm.DB) (err error) {
	if device.Model == types.SaModel {
		// sa设备已被绑定，直接返回
		if err = tx.First(&Device{}, "model = ? and area_id=?", types.SaModel, device.AreaID).Error; err == nil {
			return errors.Wrap(err, status.SaDeviceAlreadyBind)
		}

	}
	filter := Device{
		PluginID: device.PluginID,
		Identity: device.Identity,
	}

	err = tx.Where(&filter).First(&device).Error
	if errors2.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	return errors.New(status.DeviceExist)
}

func AddDevice(d *Device, tx *gorm.DB) (err error) {
	if err = CheckDeviceExist(*d, tx); err != nil {
		return
	}

	if err = tx.Unscoped().Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "identity"},
			{Name: "plugin_id"},
			{Name: "area_id"},
		},
		UpdateAll: true,
	}).Create(d).Error; err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	filter := Device{
		AreaID:   d.AreaID,
		PluginID: d.PluginID,
		Identity: d.Identity,
	}
	d.ID = 0
	if err = tx.First(d, filter).Error; err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}

	return
}

// AddSADevice 添加SA设备
func AddSADevice(device *Device, tx *gorm.DB) (err error) {
	if device.Model != types.SaModel {
		return errors2.New("invalid sa")
	}

	// 初始化角色
	err = InitRole(tx, device.AreaID)
	if err != nil {
		return err
	}

	// 创建SaCreator用户和初始化权限
	var user User
	user.AreaID = device.AreaID
	// 使用同一个db，避免发生锁数据库的问题
	if err = CreateUser(&user, tx); err != nil {
		return err
	}
	if err = SetAreaOwnerID(device.AreaID, user.ID, tx); err != nil {
		return err
	}

	return AddDevice(device, tx)
}

func GetDeviceByIdentity(identity string) (*Device, error) {
	var device Device
	if err := GetDB().Where("identity = ?", identity).First(&device).Error; err != nil {
		return nil, err
	}

	return &device, nil
}
