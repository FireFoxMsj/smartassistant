package user

import (
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/hash"
	jwt2 "github.com/zhiting-tech/smartassistant/internal/utils/jwt"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// checkQrCodeReq 扫描邀请二维码接口请求参数
type checkQrCodeReq struct {
	QrCode   string `json:"qr_code"`
	Nickname string `json:"nickname"`
	roleIds  []int
	areaId   int
}

// UserInfoResp 扫描邀请二维码接口返回数据
type UserInfoResp struct {
	UserInfo entity.UserInfo `json:"user_info"`
}

func (req *checkQrCodeReq) validateRequest(c *gin.Context) (err error) {
	if err = c.BindJSON(&req); err != nil {
		return
	}
	//	二维码是否在有效时间
	claims, err := jwt2.ValidateUserJwt(req.QrCode)
	if err != nil {
		err = errors.Wrap(err, status.QRCodeInvalid)
		return
	}

	//	二维码是否在有效时间
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		err = errors.New(status.QRCodeExpired)
		return
	}

	// 判断二维码创建者是否有生成二维码权限
	var creatorID = int(claims["uid"].(float64))
	if !entity.JudgePermit(creatorID, types.AreaGetCode) {
		err = errors.New(status.QRCodeCreatorDeny)
		return
	}

	req.areaId = int(claims["area_id"].(float64))
	// 对应家庭未删除
	_, err = entity.GetAreaByID(req.areaId)
	if err != nil {
		return
	}
	// 角色未被删除
	req.roleIds = getRoleIds(claims)
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
		resp UserInfoResp
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

func (req *checkQrCodeReq) checkQrCode(c *gin.Context) (resp UserInfoResp, err error) {
	u := session.GetUserByToken(c)

	var uRoles []entity.UserRole

	var user entity.User
	if u == nil {
		// 未加入该家庭
		user = entity.User{
			Nickname:  req.Nickname,
			Token:     hash.GetSaToken(),
			CreatedAt: time.Now(),
		}
		if err = entity.CreateUser(&user); err != nil {
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
		if err = entity.DelUserRoleByUid(u.UserID, entity.GetDB()); err != nil {
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
		Token:         user.Token,
		AccountName:   user.AccountName,
		Nickname:      user.Nickname,
		IsSetPassword: user.Password != "",
		Phone:         user.Phone,
	}

	return
}

// getRoleIds 获取角色id数组
func getRoleIds(claims jwt.MapClaims) []int {
	var roleIds []int
	if ids, ok := claims["role_ids"].(interface{}); ok {
		if rIds, ok := ids.([]interface{}); ok {
			for _, id := range rIds {
				roleIds = append(roleIds, int(id.(float64)))
			}
		}
	}
	return roleIds
}
