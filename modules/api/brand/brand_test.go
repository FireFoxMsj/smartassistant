package brand

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zhiting-tech/smartassistant/modules/api/test"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

func TestBrand(t *testing.T) {
	cases := []test.ApiTestCase{
		// type=0 请求获取所有品牌
		{
			Method: "GET",
			Path:   "/brands?type=0",
			Status: 0,
			IsArray: []string{
				"data.brands",
				"data.brands.0.plugins",
			},
		},
		// type=1 请求获取已安装品牌
		{
			Method: "GET",
			Path:   "/brands?type=1",
			Status: 0,
		},
		// type=a 错误请求
		{
			Method: "GET",
			Path:   "/brands?type=a",
			Status: errors.BadRequest,
		},
		{
			Method: "GET",
			Path:   "/brands/yeelight",
			Status: 0,
		},
		// 插件存在
		{
			Method: "GET",
			Path:   "/plugins/yeelight",
			Status: 0,
		},
		// 插件不存在
		{
			Method: "GET",
			Path:   "/plugins/hhh",
			Status: status.PluginDomainNotExist,
		},
	}

	// 先在当前目录下复制一份plugins.json文件，测试完成后再将其删除
	curDir, _ := os.Getwd()
	fmt.Println(curDir)
	rootDir := strings.Replace(curDir, filepath.Join("internal", "api", "brand"), "", 1)
	fmt.Println(rootDir)
	pj, err := ioutil.ReadFile(rootDir + "plugins.json")
	if err == nil {
		e := ioutil.WriteFile("./plugins.json", pj, 0666)
		if e == nil {
			fmt.Println("生成临时plugins.json文件成功...")
		}
	}

	// 启动插件管理
	m := plugin.GetGlobalManager()
	m.LoadPlugins()

	test.RunApiTest(t, RegisterBrandRouter, cases, test.WithRoles("管理员"))

	re := os.Remove("./plugins.json")
	if re == nil {
		fmt.Println("成功删除临时文件...")
	}
}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
