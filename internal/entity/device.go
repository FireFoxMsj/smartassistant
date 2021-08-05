package entity

import (
	errors2 "errors"
	"time"

	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/url"

	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/internal/types"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

const (
	Light  = "light"  // 灯
	Switch = "switch" // 开关
	Plug   = "plug"   // 插座
)

// Device 识别的设备
type Device struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`                     // 地址
	Identity     string    `json:"identity" gorm:"uniqueIndex"` // 设备唯一值
	Model        string    `json:"model"`                       // 型号
	SwVersion    string    `json:"sw_version"`                  // 软件版本
	Manufacturer string    `json:"manufacturer"`                // 制造商
	Type         string    `json:"type"`                        // 设备类型，如：light,switch...
	PluginID     string    `json:"plugin_id"`
	IP           string    `json:"ip"`
	Port         string    `json:"port"`
	CreatedAt    time.Time `json:"created_at"`
	LocationID   int       `json:"location_id"`
	OwnerID      int       `json:"owner_id"`
	Deleted      gorm.DeletedAt
}

func (d Device) TableName() string {
	return "devices"
}

// PluginPath 插件的地址
func (d Device) PluginPath() string {
	return url.ConcatPath("api", "plugin", d.PluginID)
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

// 获取设备，包括已删除
func GetDeviceByIDWithUnscoped(id int) (device Device, err error) {
	err = GetDB().Unscoped().First(&device, "id = ?", id).Error
	return
}

func GetDeviceByIdentity(identity string) (device Device, err error) {
	err = GetDB().First(&device, "identity = ?", identity).Error
	return
}

func GetDevices() (devices []Device, err error) {
	err = GetDB().Find(&devices).Error
	return
}

func GetDevicesByLocationID(locationId int) (devices []Device, err error) {
	err = GetDB().Order("created_at asc").Find(&devices, "location_id = ?", locationId).Error
	return
}

// IsSAOwner 是否是创建者
func IsSAOwner(userID int) bool {
	err := GetDB().Where("owner_id = ? and model = ?", userID, types.SaModel).
		First(&Device{}).Error
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

func CheckSaDeviceCreator(userID int) (err error) {
	err = GetDB().First(&Device{}, "creator_id = ?", userID).Error
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
	var newOwnerID = device.OwnerID

	if device.Model == types.SaModel {
		// sa设备已被绑定，直接返回
		if err = tx.First(&Device{}, "model = ?", types.SaModel).Error; err == nil {
			err = errors.Wrap(err, status.SaDeviceAlreadyBind)
			return
		}

	}
	err = nil
	filter := Device{
		Manufacturer: device.Manufacturer,
		Model:        device.Model,
		Identity:     device.Identity,
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
	device.OwnerID = newOwnerID
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
