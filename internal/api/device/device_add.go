package device

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/hash"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/rand"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// deviceAddReq 添加设备接口请求参数
type deviceAddReq struct {
	Device entity.Device `json:"device"` // TODO 校验
}

// deviceAddResp 添加设备接口返回数据
type deviceAddResp struct {
	ID        int             `json:"device_id"`
	PluginURL string          `json:"plugin_url"`
	UserInfo  entity.UserInfo `json:"user_info"` // 创建人的用户信息
}

// AddDevice 用于处理添加设备接口的请求
func AddDevice(c *gin.Context) {
	var (
		req         deviceAddReq
		resp        deviceAddResp
		sessionUser *session.User
		user        entity.User
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
	if req.Device.Model != types.SaModel {
		sessionUser = session.Get(c)
		if sessionUser == nil {
			err = errors.New(status.RequireLogin)
			return
		}
		if !entity.JudgePermit(sessionUser.UserID, types.DeviceAdd) {
			err = errors.New(status.Deny)
			return
		}
		req.Device.OwnerID = sessionUser.UserID

	}

	req.Device.CreatedAt = time.Now()

	if err = req.CreateDevice(); err != nil {
		return
	}

	// 设备添加成功后需要获取Creator信息
	if user, err = entity.GetUserByID(req.Device.OwnerID); err != nil {
		return
	}

	resp.UserInfo = entity.UserInfo{
		UserId:        user.ID,
		Nickname:      user.Nickname,
		IsSetPassword: user.Password != "",
		Token:         user.Token,
	}
	resp.PluginURL = DevicePluginUrl(c.Request, req.Device, user.Token)
	resp.ID = req.Device.ID
	return
}

func (req *deviceAddReq) CreateDevice() (err error) {

	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {

		if req.Device.Model == types.SaModel {
			// 初始化角色
			err = entity.InitRole(tx)
			if err != nil {
				return err
			}

			// 创建SaCreator用户和初始化权限
			if err = InitSaOwner(tx, &req.Device); err != nil {
				return err
			}
		}

		// CreateDevice 添加设备
		if err = entity.AddDevice(&req.Device, tx); err != nil {
			return err
		}

		// 添加设备为SA

		// 将权限赋给给所有角色
		var roles []entity.Role
		roles, err = entity.GetRoles()
		if err != nil {
			return err
		}
		for _, role := range roles {
			// 查看角色设备权限模板配置
			if entity.IsDeviceActionPermit(role.ID, "control") {
				var ps []types.Permission
				ps, err = ControlPermissions(req.Device)
				if err != nil {
					logrus.Error("ControlPermissionsErr:", err.Error())
					continue
				}
				role.AddPermissionsWithDB(tx, ps...)
			}
			if entity.IsDeviceActionPermit(role.ID, "update") {
				role.AddPermissionsWithDB(tx, types.NewDeviceUpdate(req.Device.ID))
			}
			if entity.IsDeviceActionPermit(role.ID, "delete") {
				role.AddPermissionsWithDB(tx, types.NewDeviceDelete(req.Device.ID))
			}
			if entity.IsDeviceActionPermit(role.ID, "manage") {
				role.AddPermissionsWithDB(tx, DeviceManagePermissions(req.Device)...)
			}
		}

		return nil
	}); err != nil {
		return
	}
	return
}

func InitSaOwner(db *gorm.DB, device *entity.Device) (err error) {
	var (
		user  entity.User
		token string
	)

	// 创建Owner用户
	user.CreatedAt = time.Now()

	token = hash.GetSaToken()
	user.Nickname = rand.String(rand.KindAll)
	user.Token = token

	if err = db.Create(&user).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	device.OwnerID = user.ID

	return
}
