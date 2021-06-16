package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.yctc.tech/root/smartassistent.git/core/http/user/utils"
	"gitlab.yctc.tech/root/smartassistent.git/utils/permission"

	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/hash"
	"gitlab.yctc.tech/root/smartassistent.git/utils/rand"
	"gitlab.yctc.tech/root/smartassistent.git/utils/response"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"
)

type updateUserReq struct {
	Nickname    *string `json:"nickname"`
	AccountName *string `json:"account_name"`
	Password    *string `json:"password"`
	RoleIds     []int   `json:"role_ids"`
}

func (req *updateUserReq) Validate(updateUid, loginId int) (updateUser orm.User, err error) {
	if len(req.RoleIds) != 0 {
		// 判断是否有修改角色权限
		if !orm.JudgePermit(loginId, permission.AreaUpdateMemberRole) {
			err = errors.Wrap(err, errors.Deny)
			return
		}
	}

	// 自己才允许修改自己的用户名,密码和昵称
	if req.Nickname != nil || req.AccountName != nil || req.Password != nil {
		if loginId != updateUid {
			err = errors.New(errors.Deny)
			return
		}
	}

	if req.Nickname != nil {
		if err = utils.CheckNickname(*req.Nickname); err != nil {
			return
		} else {
			updateUser.Nickname = *req.Nickname
		}
	}
	if req.AccountName != nil {
		if err = utils.CheckAccountName(*req.AccountName); err != nil {
			return
		} else {
			updateUser.AccountName = *req.AccountName
		}
	}

	if req.Password != nil {
		if err = utils.CheckPassword(*req.Password); err != nil {
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

func UpdateUser(c *gin.Context) {
	var (
		err         error
		req         updateUserReq
		updateUser  orm.User
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
		err = errors.Wrap(err, errors.AccountNotExistErr)
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
		// 删除用户原有角色
		if err = orm.DelUserRoleByUid(userID, orm.GetDB()); err != nil {
			return
		}
		if err = orm.CreateUserRole(WrapURoles(userID, req.RoleIds)); err != nil {
			return
		}
	}

	if err = orm.EditUser(userID, updateUser); err != nil {
		return
	}

	return
}
