package device

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
)

// 设备列表过滤条件
type listType int

// 0:所有设备;1:有控制权限的设备
const (
	AllDevice     listType = iota // 所有设备
	ControlDevice                 // 有控制权限的设备
)

// deviceListReq 设备列表接口请求参数
type deviceListReq struct {
	Type listType `form:"type"`
}

// deviceListResp 设备列表接口返回数据
type deviceListResp struct {
	Devices []Device `json:"devices"`
}

// Device 设备信息
type Device struct {
	ID           int    `json:"id"`
	Identity     string `json:"identity"`
	Name         string `json:"name"`
	Logo         string `json:"logo"` // logo相对路径
	LogoURL      string `json:"logo_url"`
	PluginID     string `json:"plugin_id"`
	LocationID   int    `json:"location_id"`
	LocationName string `json:"location_name"`
	IsSA         bool   `json:"is_sa"`
	Control      string `json:"control"` // 控制页相对路径
	PluginURL    string `json:"plugin_url"`
	Type         string `json:"type"`
}

// ListAllDevice 用于处理设备列表接口的请求
func ListAllDevice(c *gin.Context) {

	var (
		err     error
		req     deviceListReq
		resp    deviceListResp
		devices []entity.Device
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

	sessionUser := session.Get(c)
	if sessionUser == nil {
		err = errors.New(status.RequireLogin)
		return
	}

	devices, err = entity.GetDevices(sessionUser.AreaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Devices = WrapDevices(c, devices, req.Type)
	return
}

// ListLocationDevices 用于处理房间设备列表接口的请求
func ListLocationDevices(c *gin.Context) {
	var (
		err     error
		req     deviceListReq
		resp    deviceListResp
		devices []entity.Device
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
	devices, err = entity.GetDevicesByLocationID(id)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Devices = WrapDevices(c, devices, req.Type)
	return

}

func WrapDevices(c *gin.Context, devices []entity.Device, listType listType) (result []Device) {

	u := session.Get(c)
	for _, d := range devices {

		if !u.IsOwner { // 拥有者默认拥有所有权限不是拥有者则判断控制权限
			if listType == ControlDevice && !entity.DeviceControlPermit(u.UserID, d.ID) {
				continue
			}
		}

		device := Device{
			ID:         d.ID,
			Identity:   d.Identity,
			Name:       d.Name,
			Logo:       plugin.GetGlobalClient().DeviceInfo(d).Logo,
			LogoURL:    plugin.LogoURL(c.Request, d),
			LocationID: d.LocationID,
			Type:       d.Type,
		}
		if d.Model == types.SaModel {
			device.IsSA = true
			device.PluginID = ""
		} else {
			location, _ := entity.GetLocationByID(d.LocationID)
			device.LocationName = location.Name
			device.PluginID = d.PluginID
			device.Control = plugin.RelativeControlPath(d, u.Token)
			device.PluginURL = plugin.PluginURL(d, c.Request, u.Token)
		}
		result = append(result, device)

	}

	return result
}
