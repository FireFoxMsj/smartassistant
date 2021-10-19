package entity

import (
	errors2 "errors"
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
	Address      string    `json:"address"`                                                    // 地址
	Identity     string    `json:"identity" gorm:"uniqueIndex:area_id_mau_model_identity"`     // 设备唯一值
	Model        string    `json:"model" gorm:"uniqueIndex:area_id_mau_model_identity"`        // 型号
	SwVersion    string    `json:"sw_version"`                                                 // 软件版本
	Manufacturer string    `json:"manufacturer" gorm:"uniqueIndex:area_id_mau_model_identity"` // 制造商
	Type         string    `json:"type"`                                                       // 设备类型，如：light,switch...
	PluginID     string    `json:"plugin_id"`
	CreatedAt    time.Time `json:"created_at"`
	LocationID   int       `json:"location_id"`
	Deleted      gorm.DeletedAt

	AreaID uint64 `json:"area_id" gorm:"type:bigint;uniqueIndex:area_id_mau_model_identity"`
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

// IsBelongsToUserArea 是否属于用户的家庭
func (d Device) IsBelongsToUserArea(user User) bool {
	return user.BelongsToArea(d.AreaID)
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

// IsAreaOwner 是否是area拥有者
func IsAreaOwner(userID int) bool {
	user := &User{}
	if err := GetDB().Where("id = ?", userID).First(user).Error; err != nil {
		return false
	}
	err := GetDB().Where("owner_id = ? and id = ?", userID, user.AreaID).
		First(&Area{}).Error
	return err == nil
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

// IsDeviceExist 设备是否已存在
func IsDeviceExist(device *Device, tx *gorm.DB) (err error) {
	if device.Model == types.SaModel {
		// sa设备已被绑定，直接返回
		if err = tx.First(&Device{}, "model = ? and area_id=?", types.SaModel, device.AreaID).Error; err == nil {
			err = errors.Wrap(err, status.SaDeviceAlreadyBind)
			return
		}

	}
	err = nil
	filter := Device{
		Manufacturer: device.Manufacturer,
		Model:        device.Model,
		Identity:     device.Identity,
		AreaID:       device.AreaID,
	}

	if err = tx.Unscoped().Where(&filter).First(&device).Error; err != nil {
		if !errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}
		err = nil

		return
	}

	if !device.Deleted.Valid {
		err = errors.New(status.DeviceExist)
		return err
	}

	device.Deleted = gorm.DeletedAt{}
	return
}

func AddDevice(device *Device, tx *gorm.DB) (err error) {
	if err = IsDeviceExist(device, tx); err != nil {
		return
	}

	if err = tx.Save(device).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	return

}
