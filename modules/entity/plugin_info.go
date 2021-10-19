package entity

import (
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

const (
	// StatusInstallFail 添加插件失败
	StatusInstallFail = -1

	// StatusInstallSuccess 添加插件成功
	StatusInstallSuccess = 1
)

const (
	SourceTypeDefault     = "default"     // 默认插件
	SourceTypeDevelopment = "development" // 开发者插件
)

// PluginInfo 开发者插件信息
type PluginInfo struct {
	ID       int
	AreaID   uint64 `gorm:"uniqueIndex:area_plugin"`
	Area     Area   `gorm:"constraint:OnDelete:CASCADE;"`
	PluginID string `gorm:"uniqueIndex:area_plugin"`
	Info     datatypes.JSON
	Status   int
	Version  string
	Source   string
	Brand    string
}

func (p PluginInfo) TableName() string {
	return "plugin_infos"
}

func SavePluginInfo(pi PluginInfo) (err error) {
	return GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "area_id"}, {Name: "plugin_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"info"}),
	}).Create(&pi).Error
}

// UpdatePluginStatus 更新插件状态
func UpdatePluginStatus(pluginID string, status int) (err error) {
	return GetDB().Where(PluginInfo{PluginID: pluginID}).Updates(PluginInfo{Status: status}).Error
}

func IsPluginAdd(pluginID string, areaID uint64) bool {
	var pluginInfo PluginInfo
	if err := GetDB().Where(PluginInfo{PluginID: pluginID, AreaID: areaID}).First(&pluginInfo).Error; err != nil {
		logger.Errorf("get plugin info fail: %v\n", err)
		return false
	}
	return pluginInfo.Status == StatusInstallSuccess
}

// GetDevelopPlugins 获取所有开发者插件
func GetDevelopPlugins(areaID uint64) (pis []PluginInfo, err error) {
	err = GetDB().Where(PluginInfo{AreaID: areaID, Source: SourceTypeDevelopment}).Find(&pis).Error
	err = GetDB().Find(&pis).Error
	return
}

func GetPlugin(pluginID string, areaID uint64) (plugin PluginInfo, err error) {
	var pluginInfo PluginInfo
	if err = GetDB().Where(PluginInfo{PluginID: pluginID, AreaID: areaID}).First(&pluginInfo).Error; err != nil {
		return
	}
	return
}

// DelPlugin 删除插件
func DelPlugin(pluginID string, areaID uint64) (err error) {
	err = GetDB().Where(PluginInfo{PluginID: pluginID, AreaID: areaID}).
		Delete(&PluginInfo{}).Error
	return
}
