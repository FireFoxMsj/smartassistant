package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type roleInfo struct {
	orm.RoleInfo
	Permissions *Permissions `json:"permissions,omitempty"`
	IsManager   bool         `json:"is_manager"`
}

type Permissions struct {
	Device         []Permission   `json:"device"`          // 设备权限设置
	DeviceAdvanced DeviceAdvanced `json:"device_advanced"` // 设备高级权限设置
	Area           []Permission   `json:"area"`            // 家庭权限设置
	Location       []Permission   `json:"location"`        // 区域权限设置
	Role           []Permission   `json:"role"`            // 角色权限设置
	Scene          []Permission   `json:"scene"`           // 场景权限设置
}

type DeviceAdvanced struct {
	Locations []Location `json:"locations"`
}

type Location struct {
	Name    string   `json:"name"`
	Devices []Device `json:"devices"`
}
type Device struct {
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	Permission permission.Permission `json:"permission"`
	Allow      bool                  `json:"allow"` // 是否允许
}
type roleUpdateResp struct {
}

func roleUpdate(c *gin.Context) {
	var (
		err  error
		req  roleInfo
		resp roleUpdateResp
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()
	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if err = c.BindUri(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	var r orm.Role
	if orm.IsRoleNameExist(req.Name, req.ID) {
		err = errors.Wrap(err, errors.RoleNameExist)
		return
	}
	if c.Request.Method == http.MethodPut && req.ID != 0 {
		if _, err = orm.GetRoleByID(req.ID); err != nil {
			return
		}
		r, err = orm.UpdateRole(req.ID, req.Name)
	} else {
		r, err = orm.AddRole(req.Name)
	}
	if err != nil {
		return
	}
	if req.Permissions == nil {
		return
	}

	for _, v := range req.Permissions.DeviceAdvanced.Locations {
		for _, vv := range v.Devices {
			updatePermission(r, vv.Permissions)
		}
	}
	updatePermission(r, req.Permissions.Device)
	updatePermission(r, req.Permissions.Area)
	updatePermission(r, req.Permissions.Location)
	updatePermission(r, req.Permissions.Role)
	updatePermission(r, req.Permissions.Scene)
}

func updatePermission(role orm.Role, ps []Permission) {
	for _, v := range ps {
		if v.Allow {
			role.AddPermissions(v.Permission)
		} else {
			role.DelPermission(v.Permission)
		}
	}
}
