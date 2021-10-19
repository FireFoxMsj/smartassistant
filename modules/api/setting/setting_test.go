package setting

import (
	"testing"

	"github.com/zhiting-tech/smartassistant/modules/api/test"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
)

func TestSetting(t *testing.T) {
	cases := []test.ApiTestCase{
		// 获取配置
		{
			Method: "GET",
			Path:   "/setting",
			Status: 0,
		},

		// 修改配置
		{
			Method: "PUT",
			Path:   "/setting",
			Body:   "{\"user_credential_found_setting\" : {\"user_credential_found\":true}}",
			Status: status.Deny,
		},
	}

	test.RunApiTest(t, RegisterSettingRouter, cases, test.WithRoles("管理员"))
}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
