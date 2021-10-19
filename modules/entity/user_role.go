package entity

import (
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

type UserRole struct {
	ID     int
	UserID int  `gorm:"uniqueIndex:uid_rid"`
	User   User `gorm:"constraint:OnDelete:CASCADE;"`
	RoleID int  `gorm:"uniqueIndex:uid_rid"`
	Role   Role `gorm:"constraint:OnDelete:CASCADE;"`
}

func (ur UserRole) TableName() string {
	return "user_roles"
}

func CreateUserRole(uRoles []UserRole) (err error) {
	if err = GetDB().Create(&uRoles).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

func GetRoleIdsByUid(userId int) (roleIds []int, err error) {

	if err = GetDB().Model(&UserRole{}).Where("user_id = ?", userId).Pluck("role_id", &roleIds).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

func GetRolesByUid(userId int) (roles []Role, err error) {
	if err = GetDB().Model(&Role{}).
		Joins("inner join user_roles on roles.id=user_roles.role_id").
		Where("user_roles.user_id = ?", userId).Find(&roles).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

func DelUserRoleByUid(userId int, db *gorm.DB) (err error) {
	err = db.Where("user_id=?", userId).Delete(&UserRole{}).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	return
}


func UnScopedDelURoleByUid(userID int) (err error){
	err = db.Unscoped().Where("user_id=?", userID).Delete(&UserRole{}).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	return
}