package handlers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type rolePermissionsResp struct {
	Permissions map[string]bool `json:"permissions"`
}

func (resp *rolePermissionsResp) wrap(ps []Permission, userID int) {
	if len(resp.Permissions) == 0 {
		resp.Permissions = make(map[string]bool)
	}
	for _, v := range ps {
		vp := v.Permission
		strs := []string{vp.Action, vp.Target}
		if vp.Attribute != "" {
			strs = append(strs, vp.Attribute)
		}

		p := strings.Join(strs, "_")
		resp.Permissions[p] = orm.JudgePermit(userID, vp)
	}
}

func UserPermissions(c *gin.Context) {
	var (
		err  error
		resp rolePermissionsResp
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	v := c.Param("id")
	userID, err := strconv.Atoi(v)
	if err != nil {
		return
	}

	if _, err = orm.GetUserByID(userID); err != nil {
		return
	}

	var ps Permissions
	ps, err = getPermissions()
	if err != nil {
		return
	}
	resp.wrap(ps.Device, userID)
	for _, v := range ps.DeviceAdvanced.Locations {
		for _, vv := range v.Devices {
			resp.wrap(vv.Permissions, userID)
		}
	}
	resp.wrap(ps.Area, userID)
	resp.wrap(ps.Location, userID)
	resp.wrap(ps.Role, userID)
	resp.wrap(ps.Scene, userID)
}
