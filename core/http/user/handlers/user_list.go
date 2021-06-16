package handlers

import (
	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

type ListUserResp struct {
	IsCreator bool           `json:"is_creator"`
	UserCount int            `json:"user_count"`
	Users     []orm.UserInfo `json:"users"`
}

func ListUser(c *gin.Context) {
	var (
		err         error
		users       []orm.User
		resp        ListUserResp
		sessionUser *session.User
	)

	defer func() {
		if err != nil {
			resp = ListUserResp{}
		}

		if len(resp.Users) == 0 {
			resp.Users = make([]orm.UserInfo, 0)
		}
		response.HandleResponse(c, err, &resp)
	}()

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, errors.AccountNotExistErr)
		return
	}

	resp.IsCreator = orm.IsSACreator(sessionUser.UserID)

	if users, err = orm.GetRoleUsers(); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.UserCount = len(users)
	resp.Users, err = WrapUser(users)
	return

}

func WrapUser(users []orm.User) (listUsers []orm.UserInfo, err error) {
	for _, user := range users {
		listUser, err := WrapUserInfo(user)
		if err != nil {
			return nil, err
		}
		listUsers = append(listUsers, listUser)
	}
	return
}
