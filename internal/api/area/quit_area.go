package area

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
)

// QuitArea 用于处理退出家庭接口的请求
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
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	if entity.IsSAOwner(sessionUser.UserID) {
		err = errors.Wrap(err, status.OwnerQuitErr)
		return
	}

	if areaID, err = strconv.Atoi(c.Param("id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if _, err = entity.GetAreaByID(areaID); err != nil {
		return
	}

	if userID, err = strconv.Atoi(c.Param("user_id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if userID != sessionUser.UserID {
		err = errors.New(status.Deny)
		return
	}

	if err = entity.DelUser(sessionUser.UserID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return

}
