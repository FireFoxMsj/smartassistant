package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"gitlab.yctc.tech/root/smartassistent.git/utils"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

type listType int

const (
	AllDevice     listType = iota // 所有设备
	ControlDevice                 // 有控制权限的设备
)

type deviceListReq struct {
	Type listType `form:"type"`
}
type deviceListResp struct {
	Devices []Device `json:"devices"`
}
type Device struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	LogoURL      string `json:"logo_url"`
	PluginID     string `json:"plugin_id"`
	LocationID   int    `json:"location_id"`
	LocationName string `json:"location_name"`
	IsSA         bool   `json:"is_sa"`
	PluginURL    string `json:"plugin_url"`
	Type         string `json:"type"`
}

func ListAllDevice(c *gin.Context) {

	var (
		err     error
		req     deviceListReq
		resp    deviceListResp
		devices []orm.Device
	)
	defer func() {
		if resp.Devices == nil {
			resp.Devices = make([]Device, 0)
		}
		response.HandleResponse(c, err, resp)
	}()
	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	devices, err = orm.GetDevices()
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Devices = WrapDevices(devices, session.Get(c), req.Type)
	return
}

func ListLocationDevices(c *gin.Context) {
	var (
		err     error
		req     deviceListReq
		resp    deviceListResp
		devices []orm.Device
	)
	defer func() {
		if resp.Devices == nil {
			resp.Devices = make([]Device, 0)
		}
		response.HandleResponse(c, err, resp)
	}()
	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	devices, err = orm.GetDevicesByLocationID(id)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Devices = WrapDevices(devices, session.Get(c), req.Type)
	return

}

func WrapDevices(devices []orm.Device, u *session.User, listType listType) (result []Device) {

	for _, d := range devices {

		if listType == ControlDevice && !orm.DeviceControlPermit(u.UserID, d.ID) {
			continue
		}
		plg, _ := plugin.InfoByDeviceModel(d.Model)
		di, _ := plugin.SupportedDeviceInfo[d.Model]
		device := Device{
			ID:         d.ID,
			Name:       d.Name,
			LogoURL:    di.LogoURL,
			PluginID:   plg.ID,
			LocationID: d.LocationID,
			Type:       d.Type,
		}
		if d.Model == plugin.SaModel {
			device.IsSA = true
			device.PluginID = ""
			device.LogoURL = plugin.SaLogoUrl
		} else {
			device.PluginURL = utils.DevicePluginURL(
				device.ID, plg.Name, d.Model, device.Name, u.Token)
			location, _ := orm.GetLocationByID(d.LocationID)
			device.LocationName = location.Name
		}
		result = append(result, device)

	}

	return result
}
