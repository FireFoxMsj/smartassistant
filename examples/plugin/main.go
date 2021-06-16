package plugin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core"
	"log"
)

func init() {
	log.Println("plugin2 init function called")
}

// func main() {
// 	currentPath := "/Users/lanrion/Yctc/code/gitlab.yctc.tech/root/smartassistent.git/plugins"
// 	fileInfos, _ := ioutil.ReadDir(currentPath)
// 	for _, info := range fileInfos {
// 		fmt.Println(info.Name(), info.IsDir())
// 	}
// }

type Light struct {
	Name string
	Age  int
	SA   *core.SmartAssistant
}

func Register(sa *core.SmartAssistant) {
	sa.Services.Register("plugin3", "turn_on", handle, schema)
	u := Light{
		Name: "TestName",
		Age:  100,
		SA:   sa,
	}
	sa.Services.Register("plugin3", "discover", u.discover, schema)
	sa.GinEngine.GET("plugin3/list", func(c *gin.Context) {
		c.JSON(200, &u)
	})
}

func (l *Light) discover(args core.M) error {
	// 	实现设备发现的逻辑
	fmt.Println("实现设备发现的逻辑....")
	deviceData := map[string]interface{}{"name": "台灯", "ip": "127.0.0.1"}
	return l.SA.Bus.Fire("device_discovered", deviceData)
}

func handle(args core.M) error {
	fmt.Println(args)
	return nil
}

func schema(args core.M) core.M {
	str := make([]string, 0)

	for _, arg := range args {
		fmt.Println(arg)
		str = append(str, arg.(string))
	}

	var sd = make(core.M)
	sd["res"] = "123"
	return sd
}
