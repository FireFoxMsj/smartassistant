package session

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"strconv"
	"time"
)

const sessionName = "user"

type User struct {
	UserID        int                    `json:"uid"`
	IsOwner       bool                   `json:"is_owner"`
	UserName      string                 `json:"user_name"`
	RoleID        int                    `json:"role_id"`
	Token         string                 `json:"token"`
	LoginAt       time.Time              `json:"login_at"`
	LoginDuration time.Duration          `json:"login_duration"`
	ExpiresAt     time.Time              `json:"expires_at"`
	AreaID        uint64                 `json:"area_id"`
	Option        map[string]interface{} `json:"option"`
	Key           string
}

func (u User) BelongsToArea(areaID uint64) bool {
	return u.AreaID == areaID
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
	if u, exists := c.Get("userInfo"); exists {
		return u.(*User)
	}
	var u *User
	token := c.GetHeader(types.SATokenKey)
	if token != "" {
		u = GetUserByToken(c)
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
	}
	c.Set("userInfo", u)
	return u
}

func GetUserByToken(c *gin.Context) *User {
	accessToken := c.GetHeader(types.SATokenKey)
	ti, err := oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
	if err != nil {
		return nil
	}

	uid, _ := strconv.Atoi(ti.GetUserID())
	user, _ := entity.GetUserByID(uid)

	area, err := entity.GetAreaByID(user.AreaID)
	if err != nil {
		return nil
	}

	u := &User{
		UserID:   uid,
		UserName: user.AccountName,
		Token:    accessToken,
		AreaID:   user.AreaID,
		IsOwner:  area.OwnerID == user.ID,
		Key:      user.Key,
	}
	return u
}

func init() {
	gob.Register(&User{})
}
