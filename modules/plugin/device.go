package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

// AddDevice 添加设备,并保存物模型和生成设备影子
func AddDevice(d *entity.Device, tx *gorm.DB) (err error) {
	// 获取所有属性
	das, err := GetGlobalClient().GetAttributes(*d)
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
	go GetGlobalClient().HealthCheck(*d)
	return nil
}

// RemoveDevice 删除设备,断开相关连接和回收资源
func RemoveDevice(deviceID int) (err error) {
	d, err := entity.GetDeviceByID(deviceID)
	if err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	if err = GetGlobalClient().Disconnect(d); err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}

	if err = entity.DelDeviceByID(deviceID); err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	return
}

func getShadow(d entity.Device) (shadow entity.Shadow, err error) {
	// 从设备影子中获取属性
	if err = json.Unmarshal(d.Shadow, &shadow); err != nil {
		return
	}
	return
}
func getThingModel(d entity.Device) (thingModel DeviceAttributes, err error) {
	// 从设备影子中获取属性
	if err = json.Unmarshal(d.ThingModel, &thingModel); err != nil {
		return
	}
	return
}

// UpdateShadowReported 更新设备影子属性报告值
func UpdateShadowReported(d entity.Device, attr entity.Attribute) (err error) {
	// 从设备影子中获取属性
	shadow, err := getShadow(d)
	if err != nil {
		return
	}
	shadow.UpdateReported(attr.InstanceID, attr.Attribute)
	d.Shadow, err = json.Marshal(shadow)
	if err != nil {
		return
	}
	return entity.GetDB().Save(d).Error
}

// SetAttributes 通过插件设置设备的属性
func SetAttributes(areaID uint64, pluginID, identity string, data json.RawMessage) (err error) {
	d, err := entity.GetPluginDevice(areaID, pluginID, identity)
	if err != nil {
		return
	}
	_, err = GetGlobalClient().SetAttributes(d, data)
	return
}

// GetControlAttributeByID 获取设备属性（不包括设备型号、厂商等属性）
func GetControlAttributeByID(d entity.Device, instanceID int, attr string) (attribute entity.Attribute, err error) {
	as, err := GetControlAttributes(d)
	if err != nil {
		return
	}

	for _, a := range as {
		if a.InstanceID == instanceID && a.Attribute.Attribute == attr {
			return a, nil
		}
	}
	err = fmt.Errorf("plugin %s d %s instance id %d attr  %s not found",
		d.PluginID, d.Identity, instanceID, attr)
	return
}

// GetControlAttributes 获取设备属性（不包括设备型号、厂商等属性）
func GetControlAttributes(d entity.Device) (attributes []entity.Attribute, err error) {
	das, err := getThingModel(d)
	if err != nil {
		return
	}
	for _, instance := range das.Instances {
		if instance.Type == "info" {
			continue
		}
		as := GetInstanceControlAttributes(instance)
		attributes = append(attributes, as...)
	}
	return
}

// GetInstanceControlAttributes 获取实例的控制属性
func GetInstanceControlAttributes(instance Instance) (attributes []entity.Attribute) {
	for _, attr := range instance.Attributes {

		// 仅返回能控制的属性
		if attr.Attribute.Attribute == "name" {
			continue
		}
		a := entity.Attribute{
			Attribute:  attr.Attribute,
			InstanceID: instance.InstanceId,
		}
		attributes = append(attributes, a)
	}
	return
}

// GetUserDeviceAttributes 获取用户设备的属性(包括权限)
func GetUserDeviceAttributes(areaID uint64, userID int, pluginID, identity string) (das DeviceAttributes, err error) {

	device, err := entity.GetPluginDevice(areaID, pluginID, identity)
	if err != nil {
		return
	}
	das, err = getThingModel(device)
	if err != nil {
		return
	}
	for i, instance := range das.Instances {
		for j, a := range instance.Attributes {
			if entity.IsDeviceControlPermitByAttr(userID, device.ID,
				instance.InstanceId, a.Attribute.Attribute) {
				das.Instances[i].Attributes[j].CanControl = true
			}
		}
	}

	// 判断是否在线
	if !GetGlobalClient().IsOnline(device) {
		err = errors.New(status.DeviceOffline)
		return
	}
	// 在线则直接使用设备影子中缓存的属性
	shadow, err := getShadow(device)
	if err != nil {
		return
	}
	for i, ins := range das.Instances {
		for j, attr := range ins.Attributes {
			das.Instances[i].Attributes[j].Val, err = shadow.Get(ins.InstanceId, attr.Attribute.Attribute)
			if err != nil {
				return
			}
		}
	}
	das.Online = true

	return
}

// LogoURL Logo图片地址
func LogoURL(req *http.Request, d entity.Device) string {
	if d.Model == types.SaModel {
		return url.SAImageUrl(req)
	}
	logo := url.ConcatPath(url.StaticPath(), "plugin", d.PluginID, GetGlobalClient().DeviceInfo(d).Logo)
	return url.BuildURL(logo, nil, req)
}

// PluginURL 返回设备的插件控制页url
func PluginURL(d entity.Device, req *http.Request, token string) string {
	if d.Model == types.SaModel {
		return ""
	}

	q := map[string]interface{}{
		"device_id": d.ID,
		"identity":  d.Identity,
		"model":     d.Model,
		"name":      d.Name,
		"token":     token,
		"sa_id":     config.GetConf().SmartAssistant.ID,
		"plugin_id": d.PluginID,
	}
	controlPath := url.ConcatPath(url.StaticPath(), "plugin", d.PluginID, GetGlobalClient().DeviceInfo(d).Control)
	return url.BuildURL(controlPath, q, req)
}

// RelativeControlPath 返回设备的插件控制页相对路径
func RelativeControlPath(d entity.Device, token string) string {
	if d.Model == types.SaModel {
		return ""
	}

	q := map[string]interface{}{
		"device_id": d.ID,
		"identity":  d.Identity,
		"model":     d.Model,
		"name":      d.Name,
		"token":     token,
		"sa_id":     config.GetConf().SmartAssistant.ID,
		"plugin_id": d.PluginID,
	}
	return fmt.Sprintf("%s?%s", GetGlobalClient().DeviceInfo(d).Control, url.Join(url.BuildQuery(q)))
}

// ArchiveURL 插件的前端压缩包地址
func ArchiveURL(pluginID string, req *http.Request) string {

	fileName := fmt.Sprintf("%s.zip", pluginID)
	path := url.ConcatPath(url.StaticPath(), "plugin", pluginID, "resources/archive", fileName)
	return url.BuildURL(path, nil, req)
}

// StaticURL 插件的静态文件地址
func StaticURL(pluginID, relativePath string, req *http.Request) string {
	path := url.ConcatPath(url.StaticPath(), "plugin", pluginID, relativePath)
	return url.BuildURL(path, nil, req)
}
