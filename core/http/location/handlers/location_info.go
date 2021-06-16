package handlers

import (
	device "gitlab.yctc.tech/root/smartassistent.git/core/http/device/handlers"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type infoResp struct {
	Name    string       `json:"name"`
	Devices []infoDevice `json:"devices"`
}

type infoDevice struct {
	ID        int    `json:"id"`
	LogoURL   string `json:"logo_url"`
	Name      string `json:"name"`
	IsSa      bool   `json:"is_sa"`
	PluginURL string `json:"plugin_url"`
	PluginID  string `json:"plugin_id"`
}

func InfoLocation(c *gin.Context) {
	var (
		err         error
		locationId  int
		infoDevices []infoDevice
		resp        infoResp
		location    orm.Location
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

	if location, err = orm.GetLocationByID(locationId); err != nil {
		return
	}

	if infoDevices, err = GetLocationDevice(locationId, session.Get(c)); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp.Devices = infoDevices
	resp.Name = location.Name
	return

}

func GetLocationDevice(locationId int, u *session.User) (infoDevices []infoDevice, err error) {
	var (
		devices []orm.Device
	)
	devices, err = orm.GetDevicesByLocationID(locationId)
	if err != nil {
		return
	}
	deviceInfos := device.WrapDevices(devices, u, device.AllDevice)
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
