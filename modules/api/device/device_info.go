package device

import (
	errors2 "errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// infoDeviceResp 设备详情接口返回数据
type infoDeviceResp struct {
	Device infoDevice `json:"device_info"`
}

// infoDevice 设备详情
type infoDevice struct {
	ID       int          `json:"id"`
	Name     string       `json:"name"`
	LogoURL  string       `json:"logo_url"`
	Model    string       `json:"model"`
	Location infoLocation `json:"location"`
	Plugin   infoPlugin   `json:"plugin"`

	Attributes []entity.Attribute `json:"attributes"` // 有权限的action

	Permissions Permissions `json:"permissions"`
}

// Permissions 设备权限
type Permissions struct {
	UpdateDevice bool `json:"update_device"`
	DeleteDevice bool `json:"delete_device"`
}

// infoLocation 设备所属房间详情
type infoLocation struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// infoPlugin 设备的插件详情
type infoPlugin struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	URL  string `json:"url"`
}

// InfoDevice 用于处理设备详情接口的请求
func InfoDevice(c *gin.Context) {
	var (
		err    error
		id     int
		device entity.Device
		resp   infoDeviceResp
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	id, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if device, err = entity.GetDeviceByID(id); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.NotFound)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}

	if resp.Device, err = BuildInfoDevice(device, session.Get(c), c.Request); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return

}

func BuildInfoDevice(device entity.Device, user *session.User, req *http.Request) (iDevice infoDevice, err error) {
	var (
		iLocation infoLocation
		location  entity.Location
	)
	if device.LocationID > 0 {
		if location, err = entity.GetLocationByID(device.LocationID); err != nil {
			if !errors2.Is(err, gorm.ErrRecordNotFound) {
				return
			} else {
				err = nil
			}
		} else {
			iLocation.ID = device.LocationID
			iLocation.Name = location.Name
		}
	}

	iDevice = infoDevice{
		ID:       device.ID,
		Name:     device.Name,
		Model:    device.Model,
		Location: iLocation,
		LogoURL:  plugin.LogoURL(req, device),
	}

	userID := user.UserID
	if device.Model != types.SaModel {
		iDevice.Plugin = infoPlugin{
			Name: device.PluginID,
			ID:   device.PluginID,
			URL:  plugin.PluginURL(device, req, user.Token),
		}
		iDevice.Attributes, err = getDeviceAttributes(userID, device)
		if err != nil {
			return
		}
	}
	iDevice.Permissions.DeleteDevice = entity.JudgePermit(userID,
		types.NewDeviceDelete(device.ID))
	iDevice.Permissions.UpdateDevice = entity.JudgePermit(userID,
		types.NewDeviceUpdate(device.ID))

	return
}

// getDeviceAttributes 获取设备有权限的action
func getDeviceAttributes(userID int, device entity.Device) (as []entity.Attribute, err error) {

	attributes, err := plugin.GetControlAttributes(device)
	if err != nil {
		return
	}
	up, err := entity.GetUserPermissions(userID)
	if err != nil {
		return
	}
	for _, attr := range attributes {
		if !up.IsDeviceAttrPermit(device.ID, attr) {
			continue
		}
		as = append(as, attr)
	}
	return
}
