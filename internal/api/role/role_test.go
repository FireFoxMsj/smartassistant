package role

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/zhiting-tech/smartassistant/internal/api/test"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"testing"
)

func TestRole(t *testing.T) {
	cases := []test.ApiTestCase{
		{
			Method: "GET",
			Path:   "/role_tmpl",
			Status: 0,
			IsArray: []string{
				"data.role.permissions.device",
				"data.role.permissions.area",
				"data.role.permissions.location",
				"data.role.permissions.scene",
				"data.role.permissions.role",
			},
		},
		{
			Method:  "GET",
			Path:    "/roles",
			Status:  0,
			IsArray: []string{"data.roles"},
			IsID:    []string{"data.roles.0.id"},
		},
		{
			Method: "GET",
			Path:   "/roles/1",
			Status: 0,
			IsID:   []string{"data.role.id"},
		},
		{
			Method: "GET",
			Path:   "/roles/9999",
			Status: int64(status.RoleNotExist),
		},
		{
			Method: "POST",
			Body:   "{\n  \"name\": \"test_simple_role\"\n}",
			Path:   "/roles",
			Status: 0,
		},
		{
			Method: "POST",
			Body:   "{\n  \"name\": \"role_with_role\",\n  \"permissions\": {\n    \"area\": [\n      {\n        \"permission\": {\n          \"name\": \"查看角色列表\",\n          \"action\": \"get\",\n          \"target\": \"role\",\n          \"attribute\": \"\"\n        },\n        \"allow\": true\n      }\n    ]\n  }\n}",
			Path:   "/roles",
			Status: 0,
		},
		{
			Method: "PUT",
			Body:   "{\n  \"name\": \"test_role_edit\"}",
			Path:   "/roles/3",
			Status: 0,
		},
	}
	test.RunApiTest(t, RegisterRoleRouter, cases, test.WithRoles("管理员"))
	var tr0, tr1, tr2 entity.Role
	db := entity.GetDB().Where("name=?", "test_role_edit").First(&tr0)
	assert.Empty(t, db.Error)
	assert.NotEmpty(t, tr0.ID, "test_role_edit failed")
	db = entity.GetDB().Where("name=?", "role_with_role").First(&tr1)
	assert.Empty(t, db.Error)
	assert.NotEmpty(t, tr1.ID, "role_with_role failed")
	test.RunApiTest(t, RegisterRoleRouter, cases[0:1], test.WithRoles("role_with_role"))

	cases = []test.ApiTestCase{
		{
			Method: "DELETE",
			Path:   fmt.Sprintf("/roles/%d", tr1.ID),
			Status: 0,
		},
	}
	test.RunApiTest(t, RegisterRoleRouter, cases, test.WithRoles("管理员"))
	db = entity.GetDB().Where("name=?", tr1.Name).First(&tr2)
	assert.NotEmpty(t, db.Error)
}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
