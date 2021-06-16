package handlers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
	"strconv"
	"time"
)

const expireAt = time.Minute * 10

type getInvitationCodeReq struct {
	RoleIds []int `json:"role_ids"`
	AreaId  int   `json:"area_id"`
	UserId  int   `json:"-"`
}

type getInvitationCodeResp struct {
	QRCode string `json:"qr_code"`
}

func GetInvitationCode(c *gin.Context) {
	var (
		req  getInvitationCodeReq
		err  error
		resp getInvitationCodeResp
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = req.validateRequest(c); err != nil {
		return
	}
	resp, err = req.getInvitationCode(c)
}

func (req *getInvitationCodeReq) validateRequest(c *gin.Context) (err error) {
	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if len(req.RoleIds) == 0 || req.AreaId == 0 {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	req.UserId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	// 角色是否存在
	_, err = orm.GetRolesByIds(req.RoleIds)
	if err != nil {
		return
	}

	// TODO 家庭是否绑定sa

	return

}

func (req getInvitationCodeReq) getInvitationCode(c *gin.Context) (resp getInvitationCodeResp, err error) {
	// 设置jwt token
	claims := jwt.MapClaims{
		"uid":      session.Get(c).UserID,
		"area_id":  req.AreaId,
		"role_ids": req.RoleIds,
		"exp":      time.Now().Add(expireAt).Unix(),
	}

	resp.QRCode, err = utils.GetJwt(claims)
	if err != nil {
		err = errors.Wrap(err, errors.GetQRCodeErr)
	}
	return
}
