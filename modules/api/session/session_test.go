package session

import (
	"testing"

	"github.com/zhiting-tech/smartassistant/modules/api/test"
)

func TestSession(t *testing.T) {
	cases := []test.ApiTestCase{
		// 拥有 smart-assistant-token 直接登录
		{
			Method: "POST",
			Path:   "/sessions/login",
			Body:   "{\n    \"account_name\": \"\",\n    \"password\": \"\"\n}",
			Status: 0,
		},

		// 注销
		{
			Method: "POST",
			Path:   "/sessions/logout",
			Status: 0,
		},
	}

	test.RunApiTest(t, InitSessionRouter, cases, test.WithRoles("管理员"))
}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
