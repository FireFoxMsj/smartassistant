package plugin

import (
	"encoding/json"
	"errors"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/pkg/archive"
)

// LoadPluginFromZip 从压缩包中加载插件
func LoadPluginFromZip(path string, areaID uint64) (plg Plugin, err error) {

	dstDir := filepath.Dir(path)
	// unzip file
	if err = archive.UnZip(dstDir, path); err != nil {
		return
	}

	logrus.Println(dstDir)
	dstDir, _ = filepath.Abs(dstDir)
	logrus.Println(dstDir)
	plg, err = LoadPluginInfo(dstDir)
	if err != nil {
		return
	}

	// save plugin info
	info, _ := json.Marshal(plg)
	pi := entity.PluginInfo{
		AreaID:   areaID,
		PluginID: plg.ID,
		Info:     info,
		Version:  plg.Version,
		Brand:    plg.Brand,
		Source:   entity.SourceTypeDevelopment,
	}
	if err = entity.SavePluginInfo(pi); err != nil {
		return
	}

	plg.AreaID = pi.AreaID
	// docker build
	go func() {
		var err error
		defer func() {
			os.RemoveAll(dstDir)
			var status = entity.StatusInstallSuccess
			if err != nil {
				status = entity.StatusInstallFail
			}
			if uerr := entity.UpdatePluginStatus(plg.ID, status); uerr != nil {
				logrus.Errorf("UpdatePluginStatus err: %s", uerr.Error())
			}
		}()

		_, err = BuildFromDir(dstDir, plg.ID)
		if err != nil {
			logrus.Errorf("build image err: %v\n", err)
			return
		}

		if err = plg.Up(); err != nil {
			logrus.Errorf("up image err: %v\n", err)
			return
		}
		logrus.Println("image build success")
	}()

	return
}

// LoadPluginInfo 加载插件信息
func LoadPluginInfo(path string) (plg Plugin, err error) {

	configFile, err := os.Open(path + "/config.json")
	if err != nil {
		return
	}
	defer configFile.Close()

	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &plg)

	if plg.Brand == "" || plg.ID == "" {
		err = errors.New("config err")
		return
	}
	plg.Image = docker.Image{
		Name: plg.ID,
	}
	return
}

// BuildFromDir 从源码编译镜像
func BuildFromDir(path, tag string) (imageID string, err error) {
	c := docker.GetClient()
	return c.BuildFromPath(path, tag)
}

// BuildFromTar 从源码tar压缩包中build镜像
func BuildFromTar(tarPath string) (imageID string, err error) {

	tar, err := os.Open(tarPath)
	if err != nil {
		return
	}
	defer tar.Close()

	c := docker.GetClient()
	return c.BuildFromTar(tar, "")
}
