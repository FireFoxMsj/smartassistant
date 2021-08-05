package user

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
	"strconv"
)

func TransferOwner(c *gin.Context) {
	var (
		err error
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	user := session.Get(c)
	if user == nil {
		err = errors.Wrap(err, status.RequireLogin)
		return
	}

	// 判断是否是拥有者
	if !entity.IsSAOwner(user.UserID) {
		err = errors.Wrap(err, status.Deny)
		return
	}

	newOwnerID, err := strconv.Atoi(c.Param("id"))
	if err != nil || newOwnerID == user.UserID {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	_, err = entity.GetUserByID(newOwnerID)
	if err != nil {
		return
	}

	// 转移角色
	if err = entity.GetDB().Transaction(func(tx *gorm.DB) error {
		// 删除新拥有者的用户与角色信息
		if err = tx.Where("user_id=?", newOwnerID).Delete(&entity.UserRole{}).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}

		// 更新SA的拥有者
		if err = tx.Model(&entity.Device{}).Where("model=?", types.SaModel).Update("owner_id", newOwnerID).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}

		roleManager, err := entity.GetManagerRoleWithDB(tx)
		if err != nil {
			return err
		}

		// 添加旧拥有者为管理员
		uRole := entity.UserRole{
			UserID: user.UserID,
			RoleID: roleManager.ID,
		}

		if err := tx.Create(&uRole).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}

		return nil
	}); err != nil {
		return
	}

}
