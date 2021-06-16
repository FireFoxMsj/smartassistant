package handlers

import (
	"strconv"

	"gitlab.yctc.tech/root/smartassistent.git/utils/session"

	"github.com/gin-gonic/gin"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
)

type userInfoResp struct {
	orm.UserInfo
	IsCreator bool `json:"is_creator"`
	IsSelf    bool `json:"is_self"`
}

func InfoUser(c *gin.Context) {
	var (
		err         error
		resp        userInfoResp
		user        orm.User
		userID      int
		sessionUser *session.User
	)

	defer func() {
		if err != nil {
			resp = userInfoResp{}
		}

		response.HandleResponse(c, err, &resp)
	}()

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, errors.AccountNotExistErr)
		return
	}

	userID, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if user, err = orm.GetUserByID(userID); err != nil {
		return
	}

	resp.IsCreator = orm.IsSACreator(userID)

	resp.IsSelf = userID == sessionUser.UserID
	resp.UserInfo, err = WrapUserInfo(user)
	return
}

func WrapUserInfo(user orm.User) (infoUser orm.UserInfo, err error) {
	infoUser.UserId = user.ID
	infoUser.Nickname = user.Nickname
	infoUser.IsSetPassword = user.Password != ""
	infoUser.RoleInfos, err = GetRoleInfo(user.ID)
	return
}

func GetRoleInfo(uID int) (roleInfos []orm.RoleInfo, err error) {
	roles, err := orm.GetRolesByUid(uID)
	if err != nil {
		return
	}

	for _, role := range roles {
		roleInfo := orm.RoleInfo{
			ID:   role.ID,
			Name: role.Name,
		}
		roleInfos = append(roleInfos, roleInfo)
	}
	return
}
