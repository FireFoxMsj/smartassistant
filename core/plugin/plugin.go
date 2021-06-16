package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"

	"gitlab.yctc.tech/root/smartassistent.git/core/plugin/zip"
)

const (
	pluginStorePath = "./plugins/"

	SaModel   = "smart_assistant"
	SaLogoUrl = "https://tysq2.yctc.tech/api/file/originals/id/2009113/fn/智慧中心2.png"
)

type Device struct {
	LogoURL string   `json:"logo_url" yaml:"logo_url"`
	Model   string   `json:"model"`
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

type Action struct {
	Cmd           string `yaml:"cmd"`
	Name          string `yaml:"name"`
	Attribute     string `yaml:"attribute"`
	AttributeName string `yaml:"attribute_name"`
	Action        string `yaml:"action"`
}

type Plugin struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	LogoURL        string   `json:"logo_url"`
	Version        string   `json:"version"`
	Brand          string   `json:"brand"`
	Info           string   `json:"info"`
	DownloadURL    string   `json:"download_url"`
	VisitURL       string   `json:"visit_url"`
	SupportDevices []Device `json:"support_devices" yaml:"support_devices"`

	installPath string `json:"-"`
	IsAdded     bool   `json:"is_added"`
	IsNewest    bool   `json:"is_newest"`
}

func init() {
	if _, err := os.Stat(pluginStorePath); err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(pluginStorePath, os.ModePerm); err != nil {
			log.Println("初始化插件目录失败：", err.Error())
		}
	}

	initTotalPlgs()
}

// Load 加载插件，从plugins目录中获取.so文件，依次加载
func Load() error {
	plgs, err := List()
	if err != nil {
		log.Println("[plugin] Load err", err.Error())
		return err
	}

	for _, info := range plgs {
		if err = LoadOne(info.Name); err != nil {
			log.Println("[plugin] LoadOne err", err.Error())
		}
	}
	return nil
}

var cmds = make(map[string]*exec.Cmd)

// LoadOne 加载单个插件
func LoadOne(plgName string) error {

	path := fmt.Sprint(pluginStorePath, plgName, "/start.sh")

	go func() {
		os.Chmod(path, 0777)
		cmd := exec.Command(path)
		if err := cmd.Start(); err != nil {
			log.Println(err)
			return
		}
		cmds[plgName] = cmd
		cmd.Wait()

		log.Printf("plugin %s exit\n", plgName)
	}()
	return nil
}

// List 从 pluginStorePath 扫出插件信息
func List() ([]Plugin, error) {
	var plugins []Plugin

	fileInfos, err := ioutil.ReadDir(pluginStorePath)

	if err != nil {
		log.Println("[plugin] List err", err.Error())
		return plugins, err
	}

	for _, info := range fileInfos {
		if info.IsDir() {
			installPath := fmt.Sprint(pluginStorePath, info.Name())
			configFile := fmt.Sprint(installPath, "/config.yaml")
			if bytes, err := ioutil.ReadFile(configFile); err == nil {
				p := Plugin{}
				if err = yaml.Unmarshal(bytes, &p); err == nil {
					p.installPath = installPath
					p.VisitURL = fmt.Sprint(info.Name(), "/html/index.html")
					plugins = append(plugins, p)
				}
			}
		}
	}

	return plugins, nil
}

var (
	totalPlgs           []Plugin
	SADevice            = Device{LogoURL: SaLogoUrl}
	SupportedDeviceInfo = map[string]Device{SaModel: SADevice}
)

func initTotalPlgs() {

	plgsFile, err := os.Open("plugins.json")
	if err != nil {
		log.Panic(err)
	}
	defer plgsFile.Close()

	data, err := ioutil.ReadAll(plgsFile)
	if err != nil {
		log.Panic(err)
	}

	if err = json.Unmarshal(data, &totalPlgs); err != nil {
		log.Panic(err)
	}

	// 设备logo信息，从插件中加载到内存里
	for _, plg := range totalPlgs {
		for _, d := range plg.SupportDevices {
			SupportedDeviceInfo[d.Model] = d
		}
	}
}

// TotalPlugins 返回所有插件 TODO 从配置读取假数据临时使用
func TotalPlugins() []Plugin {
	plgs := make([]Plugin, len(totalPlgs))
	copy(plgs, totalPlgs)
	return plgs
}

// compareVersion 比较插件版本 TODO
func compareVersion(new, old string) bool {
	return true
}

func plgsToMap(plgs []Plugin) map[string]Plugin {
	res := make(map[string]Plugin)
	for _, plg := range plgs {
		res[plg.ID] = plg
	}
	return res
}

// wrapPlugins 包装插件信息
func wrapPlugins(total, added []Plugin) []Plugin {
	addedMap := plgsToMap(added)
	for i, plg := range total {
		if a, ok := addedMap[plg.ID]; ok {
			total[i].IsAdded = true
			total[i].IsNewest = compareVersion(plg.Version, a.Version)
		}
	}
	return total
}

// GetPlugins 获取所有插件信息
func GetPlugins() []Plugin {
	totalPlugins := TotalPlugins()
	addedPlugins, _ := List()
	return wrapPlugins(totalPlugins, addedPlugins)
}

// Info 获取插件详情
func Info(ID string) (plg Plugin, err error) {
	plgs, err := List()
	for _, plg = range plgs {
		if plg.ID == ID {
			return
		}
	}
	return
}

// GetPlugin 获取插件
func GetPlugin(ID string) (plg Plugin) {
	plgs := TotalPlugins()
	for _, plg = range plgs {
		if plg.ID == ID {
			return
		}
	}
	return
}

// InfoByDeviceModel 根据设备名获取插件详情
func InfoByDeviceModel(model string) (plg Plugin, err error) {
	plgs, err := List()
	for _, plg = range plgs {
		for _, d := range plg.SupportDevices {
			if d.Model == model {
				return
			}
		}
	}
	return
}

// Install 插件安装，包括下载
func Install(plg Plugin) error {
	log.Printf("[plugin]Install %s with url: %s \n", plg.Name, plg.DownloadURL)
	if !checkDownloadURL(plg.DownloadURL) {
		return nil
	}
	// 下载文件
	resp, err := http.Get(plg.DownloadURL)
	if err != nil {
		log.Printf("[plugin] Install download file  err %v", err)
		return err
	}

	defer resp.Body.Close()

	tmpZIP := fmt.Sprint(pluginStorePath, plg.Name, ".zip")

	if out, err := os.Create(tmpZIP); err != nil {
		return err
	} else {
		defer out.Close()
		if _, err := io.Copy(out, resp.Body); err != nil {
			log.Printf("[plugin] Install Copy err %v", err)
			return err
		}
	}

	defer os.Remove(tmpZIP)

	os.RemoveAll(fmt.Sprint(pluginStorePath, plg.Name))
	// 解压文件
	if err := zip.New(tmpZIP, fmt.Sprint(pluginStorePath, plg.Name)).Extract(); err != nil {
		log.Printf("unzip %s err %v", tmpZIP, err)
		return err
	}

	return nil
}

// Remove 插件移除
func Remove(plg Plugin) (err error) {

	fmt.Println("remove plugin:", plg.Name)
	if cmd, ok := cmds[plg.Name]; ok {
		if e := cmd.Process.Kill(); e != nil {
			log.Printf("kill process err: %s", e.Error())
		} else {
			delete(cmds, plg.Name)
		}
	}
	path := fmt.Sprint(pluginStorePath, plg.Name, "/stop.sh")
	os.Chmod(path, 0777)
	exec.Command(path).Run()

	// 删除插件
	olgPath := fmt.Sprint(pluginStorePath, plg.Name)
	if err = os.RemoveAll(olgPath); err != nil {
		return
	}
	return nil
}

// TODO checkDownloadURL 检查url是否符合规范
func checkDownloadURL(url string) bool {
	return true
}
