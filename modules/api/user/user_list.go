package user

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// ListUserResp 成员列表接口返回数据
type ListUserResp struct {
	IsOwner   bool              `json:"is_owner"`
	UserCount int               `json:"user_count"`
	Users     []entity.UserInfo `json:"users"`
}

// ListUser 用于处理成员列表接口的请求
func ListUser(c *gin.Context) {
	var (
		err         error
		users       []entity.User
		resp        ListUserResp
		sessionUser *session.User
	)

	defer func() {
		if err != nil {
			resp = ListUserResp{}
		}

		if len(resp.Users) == 0 {
			resp.Users = make([]entity.UserInfo, 0)
		}
		response.HandleResponse(c, err, &resp)
	}()

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	resp.IsOwner = entity.IsAreaOwner(sessionUser.UserID)

	if users, err = entity.GetAreaUsers(sessionUser.AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.UserCount = len(users)
	resp.Users, err = WrapUser(users)
	return

}

func WrapUser(users []entity.User) (listUsers []entity.UserInfo, err error) {
	for _, user := range users {
		listUser, err := WrapUserInfo(user)
		if err != nil {
			return nil, err
		}
		listUsers = append(listUsers, listUser)
	}
	return
}
