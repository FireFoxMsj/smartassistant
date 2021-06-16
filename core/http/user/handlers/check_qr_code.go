package handlers

import (
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/hash"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

type checkQrCodeReq struct {
	QrCode   string `json:"qr_code"`
	Nickname string `json:"nickname"`
	RoleIds  []int
	AreaId   int
}

type UserInfoResp struct {
	UserInfo orm.UserInfo `json:"user_info"`
}

func (req *checkQrCodeReq) validateRequest(c *gin.Context) (err error) {
	if err = c.BindJSON(&req); err != nil {
		return
	}
	//	二维码是否在有效时间
	claims, err := utils.ValidateJwt(req.QrCode)
	if err != nil {
		err = errors.Wrap(err, errors.QRCodeInvalid)
		return
	}

	//	二维码是否在有效时间
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		err = errors.New(errors.QRCodeExpired)
		return
	}

	// 判断二维码创建者是否有生成二维码权限
	var creatorID = int(claims["uid"].(float64))
	if !orm.JudgePermit(creatorID, permission.AreaGetCode) {
		err = errors.New(errors.QRCodeCreatorDeny)
		return
	}

	req.AreaId = int(claims["area_id"].(float64))
	// 对应家庭未删除
	_, err = orm.GetAreaByID(req.AreaId)
	if err != nil {
		return
	}
	// 角色未被删除
	req.RoleIds = getRoleIds(claims)
	roles, err := orm.GetRolesByIds(req.RoleIds)
	if err != nil {
		return
	}

	if len(roles) == 0 {
		err = errors.New(errors.RoleNameNotExist)
		return
	}
	return
}

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

	var uRoles []orm.UserRole

	var user orm.User
	if u == nil {
		// 未加入该家庭
		user = orm.User{
			Nickname:  req.Nickname,
			Token:     hash.GetSaToken(),
			CreatedAt: time.Now(),
		}
		if err = orm.CreateUser(&user); err != nil {
			return
		}
		uRoles = WrapURoles(user.ID, req.RoleIds)
	} else {
		user, err = orm.GetUserByID(u.UserID)
		if err != nil {
			return
		}

		// 重复扫码，以最后扫码角色为主
		// 删除用户原有角色
		if err = orm.DelUserRoleByUid(u.UserID, orm.GetDB()); err != nil {
			return
		}

		uRoles = WrapURoles(user.ID, req.RoleIds)
	}
	// 给用户创建角色
	if err = orm.CreateUserRole(uRoles); err != nil {
		return
	}

	resp.UserInfo = orm.UserInfo{
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

func WrapURoles(uId int, roleIds []int) (uRoles []orm.UserRole) {
	for _, roleId := range roleIds {
		uRoles = append(uRoles, orm.UserRole{
			UserID: uId,
			RoleID: roleId,
		})
	}
	return
}
