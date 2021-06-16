package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
	"gorm.io/gorm"
	"time"
)

type DataSyncReq struct {
	Nickname string   `json:"nickname"`
	Area     AreaInfo `json:"area"`
}

type AreaInfo struct {
	Name      string         `json:"name"`
	Locations []LocationInfo `json:"locations"`
}

type LocationInfo struct {
	Name string `json:"name"`
	Sort int    `json:"sort"`
}

func DataSync(c *gin.Context) {
	var (
		err         error
		req         DataSyncReq
		count       int64
		sessionUser *session.User
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, errors.AccountNotExistErr)
		return
	}

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = orm.CheckSaDeviceCreator(sessionUser.UserID); err != nil {
		err = errors.Wrap(err, errors.NotBoundDevice)
		return
	}

	// 绑定sa后,仅允许同步一次数据
	if count, err = orm.GetAreaCount(); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	} else {
		if count != 0 {
			err = errors.Wrap(err, errors.AlreadyDataSync)
			return
		}
	}

	// 同步数据
	if err = orm.GetDB().Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(&orm.User{}).Where("id = ?", sessionUser.UserID).Update("nickname", req.Nickname).Error; err != nil {
			return err
		}

		area := orm.Area{
			Name:      req.Area.Name,
			CreatedAt: time.Now(),
		}
		if err = tx.Create(&area).Error; err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return err
		}

		for _, a := range req.Area.Locations {
			location := orm.Location{
				Name:      a.Name,
				CreatedAt: time.Now(),
				Sort:      a.Sort,
			}
			if err = tx.Create(&location).Error; err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
				return err
			}
		}

		return nil
	}); err != nil {
		err = errors.Wrap(err, errors.DataSyncFail)
	}

	return
}
