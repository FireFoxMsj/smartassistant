package plugin

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/zhiting-tech/smartassistant/modules/config"

	"io/fs"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type manager struct {
	areaID  uint64
	plugins map[string]*Plugin
	docker  *docker.Client
}

// Get 获取单个插件信息
func (m *manager) Get(id string) (*Plugin, error) {
	if plg, ok := m.plugins[id]; ok {
		return plg, nil
	}
	return nil, errors.New(status.PluginDomainNotExist)
}

func NewManager() *manager {
	area, _ := getCurrentArea()
	return &manager{area.ID, make(map[string]*Plugin), docker.GetClient()}
}

// Load 加载插件列表
func (m *manager) Load() (plugins map[string]*Plugin, err error) {

	defaultPlugins, err := m.loadDefaultPlugins()
	if err != nil {
		return
	}
	for i, plg := range defaultPlugins {
		defaultPlugins[i].Source = entity.SourceTypeDefault
		defaultPlugins[i].AreaID = m.areaID
		m.plugins[plg.ID] = &defaultPlugins[i]
	}
	customPlugins, err := entity.GetDevelopPlugins(m.areaID)
	if err != nil {
		return
	}
	for _, plg := range customPlugins {
		var pi Plugin
		json.Unmarshal(plg.Info, &pi)
		pi.Image = docker.Image{
			Name: plg.PluginID,
		}
		pi.Source = plg.Source
		if _, ok := m.plugins[plg.PluginID]; !ok {
			m.plugins[plg.PluginID] = &pi
		}
	}

	return m.plugins, nil
}

// loadDefaultPlugins 加载插件列表
func (m *manager) loadDefaultPlugins() (plugins []Plugin, err error) {
	plgsFile, err := os.Open(filepath.Join(config.GetConf().SmartAssistant.DataPath(),
		"smartassistant", "plugins.json"))
	if err != nil {
		return
	}
	defer plgsFile.Close()

	data, err := ioutil.ReadAll(plgsFile)
	if err != nil {
		return
	}
	if err = json.Unmarshal(data, &plugins); err != nil {
		return
	}
	return
}

// loadCustomPlugins 加载开发者插件列表
func (m *manager) loadCustomPlugins() (plugins []Plugin, err error) {
	customDir := "./plugins/"
	var localPluginFiles []fs.FileInfo
	localPluginFiles, err = ioutil.ReadDir(customDir)
	if err != nil {
		return
	}
	for _, fileInfo := range localPluginFiles {
		if !fileInfo.IsDir() {
			continue
		}
		var plg Plugin

		plg, err = m.loadCustomPlugin(customDir + fileInfo.Name())
		if err != nil {
			return
		}
		plugins = append(plugins, plg)
	}
	return
}

// loadCustomPlugin 加载开发者插件
func (m *manager) loadCustomPlugin(path string) (plg Plugin, err error) {
	configPath := filepath.Join(path, "config.json")
	plgFile, err := os.Open(configPath)
	if err != nil {
		return
	}
	defer plgFile.Close()

	data, err := ioutil.ReadAll(plgFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &plg)
	return
}
