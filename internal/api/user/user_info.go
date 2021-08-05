package user

import (
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// userInfoResp 用户详情接口返回数据
type userInfoResp struct {
	entity.UserInfo
	IsOwner bool `json:"is_owner"`
	IsSelf  bool `json:"is_self"`
}

// InfoUser 用于处理用户详情接口的请求
func InfoUser(c *gin.Context) {
	var (
		err         error
		resp        userInfoResp
		user        entity.User
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
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	userID, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if user, err = entity.GetUserByID(userID); err != nil {
		return
	}

	resp.IsOwner = entity.IsSAOwner(userID)

	resp.IsSelf = userID == sessionUser.UserID
	resp.UserInfo, err = WrapUserInfo(user)
	resp.AccountName = user.AccountName
	return
}

func WrapUserInfo(user entity.User) (infoUser entity.UserInfo, err error) {
	infoUser.UserId = user.ID
	infoUser.Nickname = user.Nickname
	infoUser.IsSetPassword = user.Password != ""
	infoUser.RoleInfos, err = GetRoleInfo(user.ID)
	return
}

func GetRoleInfo(uID int) (roleInfos []entity.RoleInfo, err error) {

	if entity.IsSAOwner(uID) {
		roleInfos = append(roleInfos, entity.RoleInfo{
			ID:   entity.OwnerRoleID,
			Name: entity.Owner,
		})
		return
	}

	roles, err := entity.GetRolesByUid(uID)
	if err != nil {
		return
	}

	for _, role := range roles {
		roleInfo := entity.RoleInfo{
			ID:   role.ID,
			Name: role.Name,
		}
		roleInfos = append(roleInfos, roleInfo)
	}
	return
}
