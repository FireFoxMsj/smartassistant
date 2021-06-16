package handlers

import (
	"time"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"gitlab.yctc.tech/root/smartassistent.git/utils"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

type deviceAddReq struct {
	Device orm.Device `json:"device"` // TODO 校验
}

type deviceAddResp struct {
	ID        int          `json:"device_id"`
	PluginURL string       `json:"plugin_url"`
	UserInfo  orm.UserInfo `json:"user_info"` // 创建人的用户信息
}

func AddDevice(c *gin.Context) {
	var (
		req         deviceAddReq
		resp        deviceAddResp
		sessionUser *session.User
		user        orm.User
		err         error
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	// 添加非SA设备需要判断权限
	if req.Device.Model != plugin.SaModel {
		sessionUser = session.Get(c)
		if sessionUser == nil {
			err = errors.New(errors.RequireLogin)
			return
		}
		if !orm.JudgePermit(sessionUser.UserID, permission.DeviceAdd) {
			err = errors.New(errors.Deny)
			return
		}
		req.Device.CreatorID = sessionUser.UserID

		plg, _ := plugin.InfoByDeviceModel(req.Device.Model)
		resp.PluginURL = utils.DevicePluginURL(
			req.Device.ID, plg.Name, req.Device.Model, req.Device.Name, sessionUser.Token)
	}

	req.Device.CreatedAt = time.Now()

	if err = orm.CreateDevice(&req.Device); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 设备添加成功后需要获取Creator信息
	if user, err = orm.GetUserByID(req.Device.CreatorID); err != nil {
		return
	}

	resp.UserInfo = orm.UserInfo{
		UserId:        user.ID,
		Nickname:      user.Nickname,
		IsSetPassword: user.Password != "",
		Token:         user.Token,
	}
	resp.ID = req.Device.ID
	return
}
