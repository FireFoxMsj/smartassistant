package device

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
)

type ModelDevice struct {
	Name         string `json:"name"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	Logo         string `json:"logo"`         // logo地址
	Provisioning string `json:"provisioning"` // 配置页地址
	PluginID     string `json:"plugin_id"`
}

type Type struct {
	Name    string            `json:"name"`
	Type    plugin.DeviceType `json:"type"`
	Devices []ModelDevice     `json:"devices"`
}

type Response struct {
	Types []Type `json:"types"`
}

func TypeList(c *gin.Context) {

	var (
		err  error
		resp Response
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	lightType := Type{
		Name: "灯",
		Type: plugin.TypeLight,
	}
	switchType := Type{
		Name: "开关",
		Type: plugin.TypeSwitch,
	}
	outletType := Type{
		Name: "插座",
		Type: plugin.TypeOutlet,
	}
	deviceConfigs := plugin.GetGlobalClient().DeviceConfigs()
	if err != nil {
		return
	}

	for _, d := range deviceConfigs {
		if d.Provisioning == "" { //没有配置置网页则忽略
			continue
		}

		md := ModelDevice{
			Name:         d.Name,
			Model:        d.Model,
			Logo:         plugin.StaticURL(d.PluginID, d.Logo, c.Request), // 根据配置拼接插件中的图片地址
			Provisioning: d.Provisioning,
			PluginID:     d.PluginID,
		}

		switch d.Type {
		case plugin.TypeLight:
			lightType.Devices = append(lightType.Devices, md)
		case plugin.TypeSwitch:
			switchType.Devices = append(switchType.Devices, md)
		case plugin.TypeOutlet:
			outletType.Devices = append(outletType.Devices, md)
		}
	}

	resp.Types = append(resp.Types, lightType, switchType, outletType)

}
