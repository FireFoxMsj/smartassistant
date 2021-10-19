package user

import (
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"strconv"
	"testing"

	"github.com/zhiting-tech/smartassistant/modules/api/test"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/hash"
)

func TestUser(t *testing.T) {
	test.CreateRecord(&entity.Area{Name: "demo"})
	area := test.GetAreas()

	adminCases := []test.ApiTestCase{
		// 获取用户列表
		{
			Method: "GET",
			Path:   "/users",
			Status: 0,
			IsArray: []string{
				"data.users",
			},
			IsID: []string{
				"data.users.0.user_id",
			},
		},
		// 获取用户信息
		{
			Method: "GET",
			Path:   "/users/3",
			Status: 0,
		},
		// 用户不存在
		{
			Method: "GET",
			Path:   "/users/999",
			Status: status.UserNotExist,
		},
		// 用户修改自身信息
		{
			Method: "PUT",
			Path:   "/users/3",
			Body:   "{\n  \"nickname\": \"zunhuier\"}",
			Status: 0,
		},
		// 修改其它用户
		{
			Method: "PUT",
			Path:   "/users/99",
			Body:   "{\n  \"nickname\": \"adfasf\"}",
			Status: status.UserNotExist,
		},
		// 修改nickname
		{
			Method: "PUT",
			Path:   "/users/3",
			Body:   "{\n  \"nickname\": \"abc\"}",
			Status: status.NicknameLengthLowerLimit,
		},
		{
			Method: "PUT",
			Path:   "/users/3",
			Body:   "{\n  \"nickname\": \"uahidfusghidfusihgauygfisudyhfiushdiugsygfusdoa\"}",
			Status: status.NicknameLengthUpperLimit,
		},
		{
			Method: "PUT",
			Path:   "/users/3",
			Body:   "{\n  \"nickname\": \"\"}",
			Status: status.NickNameInputNilErr,
		},
		// 修改account_name
		{
			Method: "PUT",
			Path:   "/users/3",
			Body:   "{\n  \"account_name\": \"user_account_name\"}",
			Status: status.AccountNameExist,
		},
		{
			Method: "PUT",
			Path:   "/users/3",
			Body:   "{\n  \"account_name\": \"\"}",
			Status: status.AccountNameInputNilErr,
		},
		{
			Method: "PUT",
			Path:   "/users/3",
			Body:   "{\n  \"account_name\": \"a%$#c\"}",
			Status: status.AccountNameFormatErr,
		},
		{
			Method: "PUT",
			Path:   "/users/3",
			Body:   "{\n  \"account_name\": \"szc\"}",
			Status: 0,
		},
		// 管理员角色用户删除自身
		{
			Method: "DELETE",
			Path:   "/users/3",
			Status: status.DelSelfErr,
		},
		// 管理员删除成员
		{
			Method: "DELETE",
			Path:   "/users/3",
			Status: status.DelSelfErr,
		},
		// 获取邀请二维码
		{
			Method: "POST",
			Path:   "/users/3/invitation/code",
			Body:   "{\n    \"role_ids\": [1],\n    \"area_id\": " + strconv.FormatUint(area[0].ID, 10) + ",\n    \"user_id\": 3\n}",
			Status: 0,
		},
		// 角色ID为空
		{
			Method: "POST",
			Path:   "/users/3/invitation/code",
			Body:   "{\n    \"role_ids\": [],\n    \"area_id\": 1,\n    \"user_id\": 3\n}",
			Status: errors.BadRequest,
		},
		// 获取用户权限
		{
			Method: "GET",
			Path:   "/users/3/permissions",
			Status: 0,
		},
		{
			Method: "GET",
			Path:   "/users/999/permissions",
			Status: status.UserNotExist,
		},
	}

	// 先添加一个成员角色用户和一个管理员角色用户
	test.CreateRecord(&entity.User{Nickname: "user", Token: hash.GetSaToken(), AccountName: "user_account_name", AreaID: 1})
	test.CreateRecord(&entity.UserRole{UserID: 1, RoleID: 2})
	test.CreateRecord(&entity.User{Nickname: "admin", Token: hash.GetSaToken(), AreaID: 1})
	test.CreateRecord(&entity.UserRole{UserID: 2, RoleID: 1})
	test.CreateRecord(&entity.Area{ID: 1, Name: "test_area"})

	test.RunApiTest(t, RegisterUserRouter, adminCases, test.WithRoles("管理员"), test.WithAreas(area[0].ID))

	userCases := []test.ApiTestCase{
		// 成员角色用户删除成员
		{
			Method: "DELETE",
			Path:   "/users/1",
			Status: status.Deny,
		},
	}

	test.RunApiTest(t, RegisterUserRouter, userCases, test.WithRoles("成员"))
}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
