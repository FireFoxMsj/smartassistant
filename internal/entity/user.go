package entity

import (
	errors2 "errors"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"time"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

type User struct {
	ID          int       `json:"id"`
	AccountName string    `json:"account_name"`
	Nickname    string    `json:"nickname"`
	Phone       string    `json:"phone"`
	Password    string    `json:"password"`
	Salt        string    `json:"salt"`
	Token       string    `json:"token" gorm:"uniqueIndex"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserInfo struct {
	UserId        int        `json:"user_id"`
	RoleInfos     []RoleInfo `json:"role_infos"`
	AccountName   string     `json:"account_name"`
	Nickname      string     `json:"nickname"`
	Token         string     `json:"token"`
	Phone         string     `json:"phone"`
	IsSetPassword bool       `json:"is_set_password"`
}

func (u User) TableName() string {
	return "users"
}

func CreateUser(user *User) (err error) {
	err = GetDB().Create(user).Error
	return
}

func GetRoleUsers() (users []User, err error) {
	err = GetDB().Find(&users).Error
	return
}

func GetUserByID(id int) (user User, err error) {
	err = GetDB().Model(&User{}).First(&user, "id = ?", id).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.UserNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func GetUserByToken(token string) (user User, err error) {
	err = GetDB().Model(&User{}).First(&user, "token = ?", token).Error
	return
}

func EditUser(id int, updateUser User) (err error) {
	user := &User{ID: id}
	err = GetDB().First(user).Updates(&updateUser).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.UserNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func DelUser(id int) (err error) {
	user := &User{ID: id}
	err = GetDB().First(user).Delete(user).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.UserNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

func GetUserByAccountName(accountName string) (userInfo User, err error) {
	err = GetDB().Where("account_name = ?", accountName).First(&userInfo).Error
	return
}

func IsAccountNameExist(accountName string) bool {
	_, err := GetUserByAccountName(accountName)
	return err == nil
}

func (u User) BeforeDelete(tx *gorm.DB) (err error) {
	if err = DelUserRoleByUid(u.ID, tx); err != nil {
		return
	}
	return
}
