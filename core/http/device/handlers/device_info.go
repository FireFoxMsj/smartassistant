package handlers

import (
	errors2 "errors"
	"strconv"

	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

type infoDeviceResp struct {
	Device infoDevice `json:"device_info"`
}

type infoDevice struct {
	ID       int          `json:"id"`
	Name     string       `json:"name"`
	LogoURL  string       `json:"logo_url"`
	Model    string       `json:"model"`
	Location infoLocation `json:"location"`
	Plugin   infoPlugin   `json:"plugin"`
	Actions  []Action     `json:"actions"` // 有权限的action

	Permissions Permissions `json:"permissions"`
}

type Permissions struct {
	UpdateDevice bool `json:"update_device"`
	DeleteDevice bool `json:"delete_device"`
}

type Action struct {
	Action string `json:"action"` // switch/set_bright...
	Attr   string `json:"attr"`
	Name   string `json:"name"`
	Val    string `json:"val,omitempty"`
}

type infoLocation struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type infoPlugin struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func InfoDevice(c *gin.Context) {
	var (
		err    error
		id     int
		device orm.Device
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
	if device, err = orm.GetDeviceByID(id); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.NotFound)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}

	userID := session.Get(c).UserID
	if resp.Device, err = BuildInfoDevice(device, userID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return

}

func BuildInfoDevice(device orm.Device, userID int) (iDevice infoDevice, err error) {
	var (
		iLocation infoLocation
		location  orm.Location
	)
	if device.LocationID > 0 {
		if location, err = orm.GetLocationByID(device.LocationID); err != nil {
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
	}

	di, _ := plugin.SupportedDeviceInfo[device.Model]
	plg, _ := plugin.InfoByDeviceModel(device.Model)
	iDevice.LogoURL = di.LogoURL
	iDevice.Plugin = infoPlugin{
		Name: plg.Name,
		ID:   plg.ID,
	}
	iDevice.Actions = getDeviceActions(userID, device)
	iDevice.Permissions.DeleteDevice = orm.JudgePermit(userID,
		permission.NewDeviceDelete(device.ID))
	iDevice.Permissions.UpdateDevice = orm.JudgePermit(userID,
		permission.NewDeviceUpdate(device.ID))
	return
}

// getDeviceActions 获取设备有权限的action
func getDeviceActions(userID int, device orm.Device) (actions []Action) {

	actions = make([]Action, 0)
	a := make(map[string]interface{}) // 记录相同action，避免重复权限
	for _, action := range orm.GetDeviceActions(device) {
		if _, ok := a[action.Attribute]; ok {
			continue
		}
		a[action.Attribute] = true
		if !orm.IsDeviceControlPermit(userID, device.ID, action.Attribute) {
			continue
		}
		a := Action{
			Action: action.Action,
			Attr:   action.Attribute,
			Name:   action.AttributeName,
		}
		actions = append(actions, a)
	}
	return
}
