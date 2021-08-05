package location

import (
	"github.com/zhiting-tech/smartassistant/internal/api/device"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// infoResp 房间详情接口返回数据
type infoResp struct {
	Name    string       `json:"name"`
	Devices []infoDevice `json:"devices"`
}

// infoDevice 设备信息
type infoDevice struct {
	ID        int    `json:"id"`
	LogoURL   string `json:"logo_url"`
	Name      string `json:"name"`
	IsSa      bool   `json:"is_sa"`
	PluginURL string `json:"plugin_url"`
	PluginID  string `json:"plugin_id"`
}

// InfoLocation 用于处理房间详情接口的请求
func InfoLocation(c *gin.Context) {
	var (
		err         error
		locationId  int
		infoDevices []infoDevice
		resp        infoResp
		location    entity.Location
	)
	defer func() {
		if resp.Devices == nil {
			resp.Devices = make([]infoDevice, 0)
		}
		response.HandleResponse(c, err, resp)
	}()

	locationId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if location, err = entity.GetLocationByID(locationId); err != nil {
		return
	}

	if infoDevices, err = GetLocationDevice(locationId, c); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp.Devices = infoDevices
	resp.Name = location.Name
	return

}

func GetLocationDevice(locationId int, c *gin.Context) (infoDevices []infoDevice, err error) {
	var (
		devices []entity.Device
	)
	devices, err = entity.GetDevicesByLocationID(locationId)
	if err != nil {
		return
	}
	deviceInfos := device.WrapDevices(c, devices, device.AllDevice)
	for _, di := range deviceInfos {
		infoDevices = append(infoDevices, infoDevice{
			ID:        di.ID,
			LogoURL:   di.LogoURL,
			Name:      di.Name,
			IsSa:      di.IsSA,
			PluginURL: di.PluginURL,
			PluginID:  di.PluginID,
		})
	}

	return
}
