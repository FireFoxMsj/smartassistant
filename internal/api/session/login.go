package session

import (
	errors2 "errors"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/hash"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// LoginReq 用户登录接口请求参数
type LoginReq struct {
	AccountName string `json:"account_name"`
	Password    string `json:"password"`
	Token       string
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
	}

	if err = req.validateRequest(c); err != nil {
		return
	}
	resp, err = req.login(c)

}

func (req *LoginReq) validateRequest(c *gin.Context) (err error) {
	token := c.GetHeader(types.SATokenKey)
	if token != "" {
		req.Token = token
		return
	}

	if req.AccountName == "" || req.Password == "" {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	return
}

func (req LoginReq) login(c *gin.Context) (resp LoginResp, err error) {
	var u entity.User
	if req.Token != "" {
		u, err = entity.GetUserByToken(req.Token)
		if err != nil {
			err = errors.Wrap(err, status.AccountNotExistErr)
			return
		}
	} else {
		u, err = req.loginWithCookies(c)
		if err != nil {
			return
		}
	}

	resp.UserInfo = entity.UserInfo{
		UserId:        u.ID,
		AccountName:   u.AccountName,
		Nickname:      u.Nickname,
		Phone:         u.Phone,
		Token:         u.Token,
		IsSetPassword: u.Password != "",
	}

	return
}

func (req *LoginReq) loginWithCookies(c *gin.Context) (u entity.User, err error) {
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
	// 设置session
	sessionUser := &session.User{
		UserID:   u.ID,
		UserName: u.AccountName,
		Token:    u.Token,
		LoginAt:  time.Now(),
		// TODO 过期时间从配置文件中获取
		ExpiresAt: time.Now().Add(time.Duration(86400) * time.Second),
		Option:    nil,
	}
	session.Login(c, sessionUser)
	return
}
