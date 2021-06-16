package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/utils/session"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

func DelArea(c *gin.Context) {
	var (
		id          int
		err         error
		sessionUser *session.User
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	id, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, errors.AccountNotExistErr)
		return
	}
	// 非创建者不能删除家庭
	if !orm.IsSACreator(sessionUser.UserID) {
		err = errors.Wrap(err, errors.Deny)
		return
	}

	// 校验AreaID
	if _, err = orm.GetAreaByID(id); err != nil {
		return
	}

	if err = orm.DelAreaByID(id); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return

}
