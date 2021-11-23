package user

import (
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/hash"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/rand"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// updateUserReq 修改用户接口请求参数
type updateUserReq struct {
	Nickname    *string `json:"nickname"`
	AccountName *string `json:"account_name"`
	Password    *string `json:"password"`
	RoleIds     []int   `json:"role_ids"`
}

func (req *updateUserReq) Validate(updateUid, loginId int) (updateUser entity.User, err error) {
	if len(req.RoleIds) != 0 {
		// 判断是否有修改角色权限
		if !entity.JudgePermit(loginId, types.AreaUpdateMemberRole) {
			err = errors.Wrap(err, status.Deny)
			return
		}
	}

	// 自己才允许修改自己的用户名,密码和昵称
	if req.Nickname != nil || req.AccountName != nil || req.Password != nil {
		if loginId != updateUid {
			err = errors.New(status.Deny)
			return
		}
	}

	if req.Nickname != nil {
		if err = checkNickname(*req.Nickname); err != nil {
			return
		} else {
			updateUser.Nickname = *req.Nickname
		}
	}
	if req.AccountName != nil {
		if err = checkAccountName(*req.AccountName); err != nil {
			return
		} else {
			updateUser.AccountName = *req.AccountName
		}
	}

	if req.Password != nil {
		if err = checkPassword(*req.Password); err != nil {
			return
		} else {
			salt := rand.String(rand.KindAll)
			updateUser.Salt = salt
			hashNewPassword := hash.GenerateHashedPassword(*req.Password, salt)
			updateUser.Password = hashNewPassword
		}
	}

	return
}

// UpdateUser 用于处理修改用户接口的请求
func UpdateUser(c *gin.Context) {
	var (
		err         error
		req         updateUserReq
		updateUser  entity.User
		sessionUser *session.User
		userID      int
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if userID, err = strconv.Atoi(c.Param("id")); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	sessionUser = session.Get(c)
	if sessionUser == nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if updateUser, err = req.Validate(userID, sessionUser.UserID); err != nil {
		return
	}

	if len(req.RoleIds) != 0 {
		if entity.IsOwner(userID) {
			err = errors.New(status.NotAllowModifyRoleOfTheOwner)
			return
		}
		// 删除用户原有角色
		if err = entity.UnScopedDelURoleByUid(userID); err != nil {
			return
		}
		if err = entity.CreateUserRole(wrapURoles(userID, req.RoleIds)); err != nil {
			return
		}
	}

	if err = entity.EditUser(userID, updateUser); err != nil {
		return
	}

	return
}
