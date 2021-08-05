package user

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/config"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	jwt2 "github.com/zhiting-tech/smartassistant/internal/utils/jwt"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
	"time"
)

// 邀请二维码过期时间
const expireAt = time.Minute * 10

// getInvitationCodeReq 获取邀请二维码接口请求参数
type getInvitationCodeReq struct {
	RoleIds []int `json:"role_ids"`
	AreaId  int   `json:"area_id"`
	UserId  int   `json:"-"`
}

// getInvitationCodeResp 获取邀请二维码接口返回数据
type getInvitationCodeResp struct {
	QRCode string `json:"qr_code"`
}

// GetInvitationCode 用于处理获取邀请二维码接口的请求
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
	_, err = entity.GetRolesByIds(req.RoleIds)
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
		"sa_id":    config.GetConf().SmartAssistant.ID,
	}

	resp.QRCode, err = jwt2.GenerateUserJwt(claims, session.Get(c))
	if err != nil {
		err = errors.Wrap(err, status.GetQRCodeErr)
	}
	return
}
