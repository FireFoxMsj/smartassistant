package orm

import (
	errors2 "errors"
	"time"

	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gorm.io/gorm"

	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
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
	LogoURL      string    `json:"logo_url"`
	PluginID     string    `json:"plugin_id"`
	IP           string    `json:"ip"`
	Port         string    `json:"port"`
	CreatedAt    time.Time `json:"created_at"`
	LocationID   int       `json:"location_id"`
	CreatorID    int       `json:"creator_id"`
	Deleted      gorm.DeletedAt
}

func (d Device) TableName() string {
	return "devices"
}

func (d *Device) BeforeCreate(tx *gorm.DB) (err error) {
	// SA设备绑定
	if d.Model == plugin.SaModel {

		// sa设备已被绑定，直接返回
		if err = tx.First(&Device{}, "model = ?", plugin.SaModel).Error; err == nil {
			err = errors.Wrap(err, errors.AlreadyBind)
			return
		}

		// 初始化角色
		err = InitRole(tx)
		if err != nil {
			return
		}

		// 创建SaCreator用户和初始化权限
		if err = InitSaCreator(tx, d); err != nil {
			return
		}

	}
	return
}

func (d *Device) AfterCreate(tx *gorm.DB) (err error) {

	if d.Model == plugin.SaModel {
		return
	}

	// 将权限赋给给所有角色
	roles, err := GetRoles()
	if err != nil {
		return err
	}
	for _, role := range roles {
		// 查看角色设备权限模板配置
		if IsDeviceActionPermit(role.ID, "control") {
			role.AddPermissionsWithDB(tx, DeviceControlPermissions(*d)...)
		}
		if IsDeviceActionPermit(role.ID, "update") {
			role.AddPermissionsWithDB(tx, permission.NewDeviceUpdate(d.ID))
		}
		if IsDeviceActionPermit(role.ID, "delete") {
			role.AddPermissionsWithDB(tx, permission.NewDeviceDelete(d.ID))
		}
	}
	return nil
}

func GetDeviceActions(device Device) []plugin.Action {
	plg, _ := plugin.Info(device.PluginID)
	for _, d := range plg.SupportDevices {
		if d.Model != device.Model {
			continue
		}
		return d.Actions
	}
	return nil
}

// GetDeviceActionByAttr 根据设备id和属性获取action
func GetDeviceActionByAttr(device Device, attr string) plugin.Action {
	actions := GetDeviceActions(device)
	for _, action := range actions {
		if action.Attribute == attr {
			return action
		}
	}
	return plugin.Action{}
}

// DeviceControlPermissions 根据配置获取设备所有控制权限
func DeviceControlPermissions(d Device) []permission.Permission {
	actions := GetDeviceActions(d)
	var as []permission.Attribute
	a := make(map[string]interface{}) // 记录相同属性，避免重复权限
	for _, action := range actions {
		if _, ok := a[action.Attribute]; ok {
			continue
		}
		a[action.Attribute] = true
		as = append(as, permission.NewAttr(action.Attribute, action.Name))
	}
	return permission.NewDeviceControl(d.ID, as...)
}

// DevicePermissions 根据配置获取设备所有权限
func DevicePermissions(d Device) []permission.Permission {
	ps := DeviceControlPermissions(d)
	ps = append(ps, permission.NewDeviceUpdate(d.ID))
	// 非SA设备可配置删除设备权限
	if d.Model != plugin.SaModel {
		ps = append(ps, permission.NewDeviceDelete(d.ID))
	}
	return ps
}

// IsCmdPermit 控制设备的websocket命令 是否有权限
func IsCmdPermit(userID int, device Device, cmd string) bool {
	plg, _ := plugin.Info(device.PluginID)
	for _, d := range plg.SupportDevices {
		if d.Model != device.Model {
			continue
		}
		for _, action := range d.Actions {
			if action.Cmd == cmd {
				return IsDeviceControlPermit(userID, device.ID, action.Attribute)
			}
		}
	}
	// 没有配置的命令默认有权限
	return true
}

func (d *Device) AfterDelete(tx *gorm.DB) (err error) {
	// 删除设备所有相关权限
	target := permission.DeviceTarget(d.ID)
	return tx.Delete(&RolePermission{}, "target = ?", target).Error
}

func CreateDevice(device *Device) error {
	return GetDB().Transaction(func(tx *gorm.DB) error {
		var (
			err          error
			newCreatorID int
			filter       = Device{
				Manufacturer: device.Manufacturer,
				Model:        device.Model,
				Identity:     device.Identity,
			}
		)
		newCreatorID = device.CreatorID
		if err = tx.Unscoped().Where(&filter).First(&device).Error; err != nil {
			if !errors2.Is(err, gorm.ErrRecordNotFound) {
				err = errors.Wrap(err, errors.InternalServerErr)
				return err
			}
			err = nil
			if err = tx.Create(device).Error; err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
			}
			return err
		}
		if !device.Deleted.Valid {
			err = errors.New(errors.DeviceExists)
			return err
		}
		device.CreatorID = newCreatorID
		device.Deleted = gorm.DeletedAt{}
		if err = tx.Save(&device).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}
		if err = device.AfterCreate(tx); err != nil {
			return err
		}
		return err
	})

}

func GetDeviceByID(id int) (device Device, err error) {
	err = GetDB().First(&device, "id = ?", id).Error
	return
}

func GetDeletedDeviceByID(id int) (device Device, err error) {
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

// IsSACreator 是否是创建者
func IsSACreator(userID int) bool {
	err := GetDB().Where("creator_id = ? and model = ?", userID, plugin.SaModel).
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
			err = errors.Wrap(err, errors.DeviceNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func GetSaDevice() (device Device, err error) {
	err = GetDB().First(&device, "model = ?", plugin.SaModel).Error
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
