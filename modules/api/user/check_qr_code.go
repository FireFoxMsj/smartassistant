package user

import (
	"github.com/zhiting-tech/smartassistant/modules/api/area"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	jwt2 "github.com/zhiting-tech/smartassistant/modules/utils/jwt"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// checkQrCodeReq 扫描邀请二维码接口请求参数
type checkQrCodeReq struct {
	QrCode   string `json:"qr_code"`
	Nickname string `json:"nickname"`
	roleIds  []int
	areaId   uint64
}

// CheckQrCodeResp 扫描邀请二维码接口返回数据
type CheckQrCodeResp struct {
	UserInfo entity.UserInfo `json:"user_info"`
	AreaInfo area.Area       `json:"area_info"`
}

func (req *checkQrCodeReq) validateRequest(c *gin.Context) (err error) {
	if err = c.BindJSON(&req); err != nil {
		return
	}

	//	二维码是否在有效时间
	claims, err := jwt2.ValidateUserJwt(req.QrCode)
	if err != nil {
		//	二维码是否在有效时间
		if err.Error() == jwt2.ErrTokenIsExpired.Error() {
			return errors.New(status.QRCodeExpired)
		}
		err = errors.Wrap(err, status.QRCodeInvalid)
		return
	}

	// 判断是否是拥有者
	u := session.Get(c)
	if u != nil {
		if entity.IsOwner(u.UserID) {
			err = errors.New(status.OwnerForbidJoinAreaAgain)
			return
		}
	}

	// 判断二维码创建者是否有生成二维码权限
	var creatorID = claims.UID
	if !entity.JudgePermit(creatorID, types.AreaGetCode) {
		err = errors.New(status.QRCodeCreatorDeny)
		return
	}

	req.areaId = claims.AreaID
	// 对应家庭未删除
	_, err = entity.GetAreaByID(req.areaId)
	if err != nil {
		return
	}
	// 角色未被删除
	req.roleIds = claims.RoleIds
	roles, err := entity.GetRolesByIds(req.roleIds)
	if err != nil {
		return
	}

	if len(roles) == 0 {
		err = errors.New(status.RoleNotExist)
		return
	}
	return
}

// CheckQrCode 用于处理扫描邀请二维码接口的请求
func CheckQrCode(c *gin.Context) {
	var (
		req  checkQrCodeReq
		err  error
		resp CheckQrCodeResp
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = req.validateRequest(c); err != nil {
		return
	}

	resp, err = req.checkQrCode(c)
	if err != nil {
		return
	}

}

func (req *checkQrCodeReq) checkQrCode(c *gin.Context) (resp CheckQrCodeResp, err error) {
	u := session.GetUserByToken(c)

	var uRoles []entity.UserRole

	var user entity.User
	if u == nil {
		// 未加入该家庭
		user = entity.User{
			Nickname: req.Nickname,
			AreaID:   req.areaId,
		}
		if err = entity.CreateUser(&user, entity.GetDB()); err != nil {
			return
		}
		uRoles = wrapURoles(user.ID, req.roleIds)
	} else {
		user, err = entity.GetUserByID(u.UserID)
		if err != nil {
			return
		}

		// 重复扫码，以最后扫码角色为主
		// 删除用户原有角色
		if err = entity.UnScopedDelURoleByUid(u.UserID); err != nil {
			return
		}

		uRoles = wrapURoles(user.ID, req.roleIds)
	}
	// 给用户创建角色
	if err = entity.CreateUserRole(uRoles); err != nil {
		return
	}

	resp.UserInfo = entity.UserInfo{
		UserId:        user.ID,
		AccountName:   user.AccountName,
		Nickname:      user.Nickname,
		IsSetPassword: user.Password != "",
		Phone:         user.Phone,
	}

	if u != nil {
		resp.UserInfo.Token = u.Token
	} else {
		resp.UserInfo.Token, err = oauth.GetSAUserToken(user, c.Request)
		if err != nil {
			return
		}

	}

	resp.AreaInfo = area.Area{
		ID: strconv.FormatUint(req.areaId, 10),
	}

	return
}
