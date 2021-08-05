package session

import (
	"encoding/gob"
	"fmt"
	orm2 "github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types"
	"time"

	"github.com/gin-gonic/gin"
)

const sessionName = "user"

type User struct {
	UserID        int                    `json:"uid"`
	UserName      string                 `json:"user_name"`
	RoleID        int                    `json:"role_id"`
	Token         string                 `json:"token"`
	LoginAt       time.Time              `json:"login_at"`
	LoginDuration time.Duration          `json:"login_duration"`
	ExpiresAt     time.Time              `json:"expires_at"`
	Option        map[string]interface{} `json:"option"`
}

func Login(c *gin.Context, user *User) {
	s := GetSession(c)
	s.Set(sessionName, user)
	if err := s.Save(); err != nil {
		fmt.Errorf("save session err: %s", err)
	}
}

func Logout(c *gin.Context) {
	s := GetSession(c)
	s.Delete(sessionName)
	if err := s.Save(); err != nil {
		fmt.Errorf("save session err: %s", err)
	}
}

// Get 根据token或cookie获取用户数据
func Get(c *gin.Context) *User {
	var u *User
	token := c.GetHeader(types.SATokenKey)
	if token != "" {
		return GetUserByToken(c)
	} else {
		// token 为空，则检查cookie
		s := GetSession(c)
		user := s.Get(sessionName)
		if user == nil {
			return nil
		}

		u = user.(*User)
		if u.UserID == 0 {
			return nil
		}
		if time.Now().After(u.ExpiresAt) || time.Now().Before(u.LoginAt) {
			return nil
		}
		return u
	}
}

func GetUserByToken(c *gin.Context) *User {
	token := c.GetHeader(types.SATokenKey)
	user, err := orm2.GetUserByToken(token)
	if err != nil {
		return nil
	}
	u := &User{
		UserID:   user.ID,
		UserName: user.AccountName,
		Token:    token,
	}
	return u
}

func init() {
	gob.Register(&User{})
}
