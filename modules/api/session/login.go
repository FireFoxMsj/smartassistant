package session

import (
	errors2 "errors"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/hash"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
	"time"
)

// LoginReq 用户登录接口请求参数
type LoginReq struct {
	AccountName string `json:"account_name" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

// LoginResp 用户登录接口返回数据
type LoginResp struct {
	UserInfo entity.UserInfo `json:"user_info"`
}

// Login 用于处理用户登录的请求
func Login(c *gin.Context) {
	var (
		req  LoginReq
		resp LoginResp
		err  error
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	resp, err = req.login(c)

}

func (req LoginReq) login(c *gin.Context) (resp LoginResp, err error) {
	var (
		u     entity.User
		token string
	)

	u, token, err = req.loginWithCookies(c)
	if err != nil {
		return
	}

	resp.UserInfo = entity.UserInfo{
		UserId:        u.ID,
		AccountName:   u.AccountName,
		Nickname:      u.Nickname,
		Phone:         u.Phone,
		Token:         token,
		IsSetPassword: u.Password != "",
	}

	return
}

func (req *LoginReq) loginWithCookies(c *gin.Context) (u entity.User, token string, err error) {
	// 判断是否存在该用户
	u, err = entity.GetUserByAccountName(req.AccountName)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.AccountNotExistErr)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	// 校验密码是否正确
	if !hash.CheckPassword(req.Password, u.Salt, u.Password) {
		err = errors.New(status.AccountPassWordErr)
		return
	}
	area, err := entity.GetAreaByID(u.AreaID)
	if err != nil {
		return
	}

	// TODO 用户Token使用oauth2生成，后续删除登录接口
	token, err = oauth.GetSAUserToken(u, c.Request)
	if err != nil {
		return
	}

	// 设置session
	sessionUser := &session.User{
		UserID:   u.ID,
		IsOwner:  area.OwnerID == u.ID,
		UserName: u.AccountName,
		Token:    token,
		LoginAt:  time.Now(),
		// TODO 过期时间从配置文件中获取
		ExpiresAt: time.Now().Add(time.Duration(86400) * time.Second),
		AreaID:    u.AreaID,
		Option:    nil,
		Key:       u.Key,
	}
	session.Login(c, sessionUser)
	return
}
