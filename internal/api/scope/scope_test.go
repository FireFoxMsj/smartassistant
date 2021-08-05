package scope

import (
	"github.com/zhiting-tech/smartassistant/internal/api/test"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"testing"
)

func TestScope(t *testing.T) {
	cases := []test.ApiTestCase{
		// 获取范围权限列表
		{
			Method: "GET",
			Path:   "/scopes",
			Status: 0,
			IsArray: []string{
				"data.scopes",
			},
		},
		// 获取 JWT
		{
			Method: "POST",
			Path:   "/scopes/token",
			Body:   "{\n    \"scopes\": [\"\"]\n}",
			Status: errors.BadRequest,
		},
		{
			Method: "POST",
			Path:   "/scopes/token",
			Body:   "{\n    \"scopes\": [\"hhh\"]\n}",
			Status: errors.BadRequest,
		},
		{
			Method: "POST",
			Path:   "/scopes/token",
			Body:   "{\n    \"scopes\": [\"user\"]\n}",
			Status: 0,
		},
		{
			Method: "POST",
			Path:   "/scopes/token",
			Body:   "{\n    \"scopes\": [\"user\",\"area\"]\n}",
			Status: 0,
		},
	}

	test.RunApiTest(t, RegisterScopeRouter, cases, test.WithRoles("管理员"))
}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
