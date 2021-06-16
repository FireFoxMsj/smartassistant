package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
	"strconv"
)

func DelUser(c *gin.Context) {
	var (
		err         error
		userID      int
		sessionUser *session.User
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	userID, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if _, err = orm.GetUserByID(userID); err != nil {
		return
	}

	if orm.IsSACreator(userID) {
		err = errors.New(errors.Deny)
		return
	}

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, errors.AccountNotExistErr)
		return
	}

	// 成员本人不能删除自己
	if sessionUser.UserID == userID {
		err = errors.Wrap(err, errors.DelSelfErr)
		return
	}

	if err = orm.DelUser(userID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	return

}
