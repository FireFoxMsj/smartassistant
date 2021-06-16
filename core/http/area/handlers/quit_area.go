package handlers

import (
	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
	"strconv"
)

func QuitArea(c *gin.Context) {
	var (
		err         error
		sessionUser *session.User
		userID      int
		areaID      int
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, errors.AccountNotExistErr)
		return
	}

	if orm.IsSACreator(sessionUser.UserID) {
		err = errors.Wrap(err, errors.CreatorQuitErr)
		return
	}

	if areaID, err = strconv.Atoi(c.Param("id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if _, err = orm.GetAreaByID(areaID); err != nil {
		return
	}

	if userID, err = strconv.Atoi(c.Param("user_id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if userID != sessionUser.UserID {
		err = errors.New(errors.Deny)
		return
	}

	if err = orm.DelUser(sessionUser.UserID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return

}
