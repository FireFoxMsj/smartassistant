package device

import (
	errors2 "errors"
	"strconv"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/plugin"

	"github.com/zhiting-tech/smartassistant/modules/api/area"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gorm.io/gorm"
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
	AreaInfo  area.Area       `json:"area_info"` // 家庭信息
}

// AddDevice 用于处理添加设备接口的请求
func AddDevice(c *gin.Context) {
	var (
		req         deviceAddReq
		resp        deviceAddResp
		sessionUser *session.User
		user        entity.User
		err         error
		areaID      uint64
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
		areaID = sessionUser.AreaID
		req.Device.CreatedAt = time.Now()
		if err = req.CreateDevice(areaID); err != nil {
			return
		}
		if user, err = entity.GetUserByID(sessionUser.UserID); err != nil {
			return
		}
	}else {
		// 判断SA是否存在
		_, err = entity.GetSaDevice()
		if err == nil {
			err = errors.Wrap(err, status.SaDeviceAlreadyBind)
			return
		} else {
			if !errors2.Is(err, gorm.ErrRecordNotFound) {
				err = errors.Wrap(err, errors.InternalServerErr)
				return
			}
		}

		areaObj, err := entity.CreateArea("")
		if err != nil {
			return
		}
		areaID = areaObj.ID
		req.Device.CreatedAt = time.Now()
		if err = req.CreateDevice(areaID); err != nil {
			return
		}
		areaObj, err = entity.GetAreaByID(areaID)
		if err != nil {
			return
		}
		if user, err = entity.GetUserByID(areaObj.OwnerID); err != nil {
			return
		}
	}
	// 设备添加成功后需要获取Creator信息
	resp.UserInfo = entity.UserInfo{
		UserId:        user.ID,
		Nickname:      user.Nickname,
		IsSetPassword: user.Password != "",
		Token:         user.Token,
	}
	resp.PluginURL = plugin.PluginURL(req.Device, c.Request, user.Token)
	resp.ID = req.Device.ID

	resp.AreaInfo = area.Area{
		ID: strconv.FormatUint(req.Device.AreaID, 10),
	}
	return
}

func (req *deviceAddReq) CreateDevice(areaID uint64) (err error) {

	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {

		if req.Device.Model == types.SaModel {
			// 初始化角色
			err = entity.InitRole(tx, areaID)
			if err != nil {
				return err
			}

			// 创建SaCreator用户和初始化权限
			var user entity.User
			user.AreaID = areaID
			// 使用同一个db，避免发生锁数据库的问题
			if err = entity.CreateUser(&user, tx); err != nil {
				return err
			}
			if err = entity.SetAreaOwnerID(areaID, user.ID, tx); err != nil {
				return err
			}
		}

		req.Device.AreaID = areaID
		// CreateDevice 添加设备
		switch req.Device.Model {
		case types.SaModel:
			// 	// 添加设备为SA时不需要添加设备影子
			if err = entity.AddDevice(&req.Device, tx); err != nil {
				return err
			}
		default:
			if err = plugin.AddDevice(&req.Device, tx); err != nil {
				return err
			}
		}

		// 将权限赋给给所有角色
		var roles []entity.Role
		roles, err = entity.GetRoles(areaID)
		if err != nil {
			return err
		}
		for _, role := range roles {
			// 查看角色设备权限模板配置
			if entity.IsDeviceActionPermit(role.ID, "control") {
				var ps []types.Permission
				ps, err = ControlPermissions(req.Device)
				if err != nil {
					logger.Error("ControlPermissionsErr:", err.Error())
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
