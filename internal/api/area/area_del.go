package area

import (
	"github.com/zhiting-tech/smartassistant/internal/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// DelArea 用于处理删除家庭接口的请求
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
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}
	// 非创建者不能删除家庭
	if !entity.IsSAOwner(sessionUser.UserID) {
		err = errors.Wrap(err, status.Deny)
		return
	}

	// 校验AreaID
	if _, err = entity.GetAreaByID(id); err != nil {
		return
	}

	if err = entity.DelAreaByID(id); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	cloud.RemoveSA()
	return
}
