package supervisor

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	errors2 "github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gopkg.in/yaml.v2"
)

const (
	stageDir = "stage"
)

var (
	manager *Manager
	_once   sync.Once
	saImage = docker.Image{
		Name:     "smartassistant",
		Tag:      types.Version,
		Registry: types.DockerRegistry,
	}
)

type Manager struct {
	// 运行时目录，docker-compose.yaml 所在
	RuntimePath string
	BackupPath  string
}

func GetManager() *Manager {
	_once.Do(func() {
		ab, err := filepath.Abs(config.GetConf().SmartAssistant.BackupPath())
		if err != nil {
			logger.Errorf("backup path not valid: %v", err)
		}
		ar, err := filepath.Abs(config.GetConf().SmartAssistant.RuntimePath)
		if err != nil {
			logger.Errorf("runtime path not valid: %v", err)
		}
		manager = &Manager{
			RuntimePath: ar,
			BackupPath:  ab,
		}
		_ = os.MkdirAll(manager.BackupPath, os.ModePerm)
		f, err := os.Stat(manager.BackupPath)
		if os.IsNotExist(err) {
			logger.Errorf("can not create backup path %v", manager.BackupPath)
		}
		if !f.IsDir() {
			logger.Error("backup path is not a dir")
		}
	})
	return manager
}

func (m *Manager) ListBackups() []Backup {
	bks := make([]Backup, 0)
	filepath.Walk(m.BackupPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// TODO 只查找第一层的zip文件
		if info.IsDir() {
			return nil
		}
		if filepath.Dir(path) != m.BackupPath {
			return nil
		}
		if filepath.Ext(path) != ".zip" {
			return nil
		}
		bks = append(bks, NewBackupFromFileName(info.Name()))
		return nil
	})
	return bks
}

// ProcessBackupJob 处理备份，恢复功能
func (m *Manager) ProcessBackupJob() (err error) {
	stage, err := loadStage(filepath.Join(m.BackupPath, stageDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		logger.Warnf("read stage file error %v", err)
		return
	}
	logger.Infof("processing backup job (%v)", stage.Value)
	defer func() {
		// 无论是否成功均需要删除过程文件
		if err = stage.remove(); err != nil {
			logger.Warnf("can not remove stage file %v", err)
		} else {
			logger.Info("remove stage file ok")
		}
	}()
	switch stage.Value {
	case StageBackupInit:
		logger.Infof("start backup %v", stage.Note)
		backup, err := m.Backup(stage.Note)
		if err != nil {
			return err
		}
		logger.Infof("backup success: %v", backup.Note)
	case StageRestoreInit:
		logger.Infof("start restore %v", stage.Note)
		err = m.Restore(stage.Note)
		return err
	}
	return
}

func stopAllPlugins() (err error) {
	resumeContainer := func(plgs []plugin.Plugin) {
		for _, plg := range plgs {
			_, _ = plugin.RunPlugin(plg)
		}
	}
	plgs, err := plugin.GetGlobalManager().LoadPlugins()
	cli := docker.GetClient()
	if err != nil {
		return
	}
	stoppedPlgs := make([]plugin.Plugin, 0)
	for _, plg := range plgs {
		ps, _ := docker.GetClient().ContainerIsRunningByImage(plg.Image)
		if ps == false {
			continue
		}

		err = cli.ContainerStopByImage(plg.Image)
		if err != nil {
			resumeContainer(stoppedPlgs)
			return err
		} else {
			stoppedPlgs = append(stoppedPlgs, *plg)
		}
	}
	return
}

func startAllPlugins() (err error) {
	plgs, err := plugin.GetGlobalManager().LoadPlugins()
	if err != nil {
		return
	}
	for _, plg := range plgs {
		plugin.RunPlugin(*plg)
	}
	return
}

func (m *Manager) processRestart(cn string) (err error) {
	err = stopAllPlugins()
	if err != nil {
		return
	}
	// 返回操作结果后重启
	go func() {
		time.Sleep(time.Second)
		err = docker.GetClient().DockerClient.ContainerRestart(context.Background(),
			cn, nil)
		if err != nil {
			logger.Warnf("restart self error %v", err)
		}
	}()
	return
}

// StartBackupJob 开始备份，将创建过程文件，关闭所有插件，然后重启
func (m *Manager) StartBackupJob(note string) (err error) {
	id, err := docker.GetClient().GetContainerByImage(saImage.RefStr())
	if err != nil {
		logger.Warnf("cannot find container %v, %v", saImage.RefStr(), err)
		return
	}
	s := NewStage(filepath.Join(m.BackupPath, stageDir), StageBackupInit)
	s.Note = note
	err = s.save()
	if err != nil {
		return
	}
	err = m.processRestart(id)
	if err != nil {
		_ = s.remove()
	}
	return
}

// Backup 接着 StartBackupJob，备份镜像，备份文件，打包
func (m *Manager) Backup(note string) (backup *Backup, err error) {
	logger.Infof("creating backup (%v)", note)
	dir, err := ioutil.TempDir(m.BackupPath, "tmp")
	if err != nil {
		return
	}
	defer os.RemoveAll(dir)
	backup = newBackup(note)
	err = backup.Save(dir)
	if err != nil {
		logger.Infof("backup create with error (%v)", err)
		return
	}
	f, err := os.Create(filepath.Join(m.BackupPath, backup.FileName))
	if err != nil {
		return
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	defer zw.Close()
	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		logger.Infof("path %v", path)
		if !fi.IsDir() {
			data, err := os.Open(path)
			if err != nil {
				return err
			}
			defer data.Close()
			w, err := zw.Create(strings.TrimPrefix(path, dir))
			if err != nil {
				return err
			}
			_, err = io.Copy(w, data)
			return err
		}
		return nil
	})
	if err != nil {
		logger.Warnf("pack backup error (%v)", err)
		return
	}
	logger.Infof("backup success")
	return
}

func (m *Manager) merge() (err error) {
	return m.mergeConfigFile()
}

func (m *Manager) mergeConfigFile() (err error) {
	var (
		content []byte
	)
	file := path.Join(m.BackupPath, stageDir, "config", "smartassistant.yaml")
	tmpfile := fmt.Sprintf("%s.bak", file)
	err = os.Rename(file, tmpfile)
	if err != nil {
		return
	}

	cloudOptions := config.Options{}
	content, err = ioutil.ReadFile(tmpfile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(content, &cloudOptions)
	if err != nil {
		return
	}

	cloudOptions.SmartAssistant.ID = config.GetConf().SmartAssistant.ID
	cloudOptions.SmartAssistant.Key = config.GetConf().SmartAssistant.Key
	cloudOptions.SmartAssistant.HostRuntimePath = config.GetConf().SmartAssistant.HostRuntimePath
	cloudOptions.SmartAssistant.RuntimePath = config.GetConf().SmartAssistant.RuntimePath

	content, err = yaml.Marshal(cloudOptions)
	err = ioutil.WriteFile(file, content, os.ModePerm)
	if err == nil {
		os.Remove(tmpfile)
	}
	return
}

// StartRestoreCloudJob 启动恢复，导入镜像，将创建过程文件，然后重启
func (m *Manager) StartRestoreCloudJob(name string) (err error) {
	id, err := docker.GetClient().GetContainerByImage(saImage.RefStr())
	if err != nil {
		logger.Infof("read Container id error %v", err)
		return
	}
	fn := filepath.Join(m.BackupPath, name)
	logger.Infof("starting restore from %v", fn)
	fi, err := os.Stat(fn)
	if err != nil {
		logger.Warnf("stat error %v", err)
		return err
	}
	sDir := filepath.Join(m.BackupPath, stageDir)
	s := NewStage(sDir, StageRestoreInit)
	_ = s.remove()
	_ = os.MkdirAll(sDir, os.ModePerm)
	err = m.unzip(fn, sDir)
	os.Remove(fn)
	if err != nil {
		return
	}
	backup := loadBackup(sDir)
	if backup == nil {
		return errors.New("load backup error")
	}
	// 合并本地信息到云端备份文件
	err = m.merge()
	if err != nil {
		return
	}
	s.Note = backup.Note
	if err != nil {
		return
	}
	err = s.save()
	if err != nil {
		return
	}
	err = m.processRestart(id)
	if err != nil {
		_ = s.remove()
	}
	logger.Infof("restore from %v", fi.Name())
	return
}

// StartRestoreJob 启动恢复，导入镜像，将创建过程文件，然后重启
func (m *Manager) StartRestoreJob(name string) (err error) {
	id, err := docker.GetClient().GetContainerByImage(saImage.RefStr())
	if err != nil {
		return
	}
	fn := filepath.Join(m.BackupPath, filepath.Clean("/"+name))
	logger.Infof("starting restore from %v", fn)
	fi, err := os.Stat(fn)
	if err != nil {
		logger.Warnf("stat error %v", err)
		return err
	}
	sDir := filepath.Join(m.BackupPath, stageDir)
	s := NewStage(sDir, StageRestoreInit)
	_ = s.remove()
	_ = os.MkdirAll(sDir, os.ModePerm)
	err = m.unzip(fn, sDir)
	if err != nil {
		return
	}
	backup := loadBackup(sDir)
	if backup == nil {
		return errors.New("load backup error")
	}
	s.Note = backup.Note
	err = s.save()
	if err != nil {
		return
	}
	err = m.processRestart(id)
	if err != nil {
		_ = s.remove()
	}
	logger.Infof("restore from %v", fi.Name())
	return
}

// Restore 接着 StartRestoreJob，替换文件，导入镜像，通过 supervisor 重启
func (m *Manager) Restore(note string) (err error) {
	sDir := filepath.Join(m.BackupPath, stageDir)
	backup := loadBackup(sDir)
	if backup == nil {
		return errors.New("load backup error")
	}
	logger.Infof("creating backup (%v)", note)
	dir, err := ioutil.TempDir(m.BackupPath, "tmp")
	if err != nil {
		logger.Errorf("create backup error %v", err)
		return
	}
	defer os.RemoveAll(dir)
	bak := newBackup(fmt.Sprintf("%v-stage", note))
	err = bak.Save(dir)
	if err != nil {
		logger.Errorf("save backup error %v", err)
		return
	}
	err, needRollback := m.restoreFromDir(sDir)
	if err == nil { // 使用supervisor重启
		logger.Info("restore file ok, resting...")
		os.RemoveAll(filepath.Join(m.BackupPath, stageDir))
		os.RemoveAll(dir)
		go func() {
			err = Restart(backup.SmartAssistant.RefStr())
			if err != nil {
				logger.Errorf("restart error %v", err)
			}
		}()
		time.Sleep(10 * time.Second)
		// TODO 恢复
		logger.Warnf("restart failed...")
		return
	}
	if needRollback {
		if err2, _ := m.restoreFromDir(dir); err2 != nil {
			logger.Errorf("restore failed %v, rollback error %v", err, err2)
		}
	}
	return

}

func (m *Manager) restoreFromDir(dir string) (err error, needRollback bool) {
	if _, err = os.Stat(dir); err != nil {
		logger.Errorf("restore error %v", err)
		return
	}
	imgFile := filepath.Join(dir, "images.tar")
	_, err = os.Stat(imgFile)
	if err == nil {
		err = docker.GetClient().ImageLoad(imgFile)
		if err != nil {
			logger.Errorf("image load error %v", err)
			return
		}
	}
	// 替换文件
	fps := []string{
		"docker-compose.yaml",
		"config",
		"data",
	}
	needRollback = true
	for _, f := range fps {
		_, err = os.Stat(filepath.Join(dir, f))
		if err != nil {
			logger.Infof("file not exist, skip, error : %v", err)
			continue
		}
		err = os.RemoveAll(filepath.Join(m.RuntimePath, f))
		if err != nil {
			logger.Errorf("backup remove error %v", err)
			return
		}
		err = os.Rename(filepath.Join(dir, f), filepath.Join(m.RuntimePath, f))
		if err != nil {
			logger.Errorf("backup restore error %v", err)
			return
		}
	}
	needRollback = false
	return
}

func (m *Manager) unzip(fn, dst string) (err error) {
	zr, err := zip.OpenReader(fn)
	if err != nil {
		return
	}
	defer zr.Close()
	for _, file := range zr.File {
		p := filepath.Join(dst, file.Name)
		logger.Infof("extracting %v", file.Name)
		// 如果是目录，就创建目录
		if file.FileInfo().IsDir() {
			logger.Infof("creating dir")
			if err := os.MkdirAll(p, file.Mode()); err != nil {
				return err
			}
			continue
		}

		// 获取到 Reader
		fr, err := file.Open()
		if err != nil {
			logger.Warnf("reader err %v", err)
			return err
		}
		fdir := filepath.Dir(p)
		os.MkdirAll(fdir, os.ModePerm)
		// 创建要写出的文件对应的 Write
		fw, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			logger.Warnf("writer err %v", err)
			fr.Close()
			return err
		}

		_, err = io.Copy(fw, fr)
		fw.Close()
		fr.Close()
		if err != nil {
			logger.Warnf("copy err %v", err)
			return err
		}
	}
	return
}

func (m *Manager) DeleteBackup(fn string) error {
	fn = filepath.Join(m.BackupPath, filepath.Clean("/"+fn))
	fi, err := os.Stat(fn)
	if err != nil {
		if os.IsNotExist(err) {
			return errors2.New(status.FileNotExistErr)
		}
		return err
	}
	if fi.IsDir() {
		return errors2.New(status.FileNotExistErr)
	}
	return os.RemoveAll(fn)
}

// StartUpdateJob 下载新版镜像，通知supervisor以新镜像重启
func (m *Manager) StartUpdateJob(version string) (err error) {
	newImage := docker.Image{
		Name:     "smartassistant",
		Tag:      version,
		Registry: types.DockerRegistry,
	}
	err = docker.GetClient().Pull(newImage.RefStr())
	if err != nil {
		err = errors2.New(status.ImagePullErr)
		return
	}
	err = stopAllPlugins()
	if err != nil {
		return
	}
	go func() {
		err = Restart(newImage.RefStr())
		if err != nil {
			logger.Errorf("restart error %v", err)
		}
		time.Sleep(10 * time.Second)
		logger.Warnf("restart failed...")
	}()
	return
}
