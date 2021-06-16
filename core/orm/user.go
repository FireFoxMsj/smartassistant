package orm

import (
	errors2 "errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/hash"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"
	"gitlab.yctc.tech/root/smartassistent.git/utils/rand"
	"gorm.io/gorm"
	"time"
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
			err = errors.Wrap(err, errors.UserNotExist)
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
			err = errors.Wrap(err, errors.UserNotExist)
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
			err = errors.Wrap(err, errors.UserNotExist)
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

func InitSaCreator(db *gorm.DB, device *Device) (err error) {
	var (
		user  User
		role  Role
		token string
	)

	// 创建Creator用户
	user.CreatedAt = time.Now()
	role, err = GetManagerRoleWithDB(db)
	token = hash.GetSaToken()
	user.Nickname = rand.String(rand.KindAll)
	user.Token = token

	if err = db.Create(&user).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	device.CreatorID = user.ID

	// SA创建者默认角色为管理员
	uRole := UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}

	if err = db.Create([]UserRole{uRole}).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 将SA权限赋给创建者
	role.AddPermissionsWithDB(db, DeviceControlPermissions(*device)...)
	role.AddPermissionsWithDB(db, permission.NewDeviceUpdate(device.CreatorID))

	return
}
