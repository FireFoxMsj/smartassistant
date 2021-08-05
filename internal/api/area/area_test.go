package area

import (
	"github.com/zhiting-tech/smartassistant/internal/api/test"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"testing"
)

func TestArea(t *testing.T) {
	userCases := []test.ApiTestCase{
		{
			Method: "GET",
			Path:   "/areas",
			Status: 0,
			IsArray: []string{
				"data.areas",
			},
		},
		{
			Method: "PUT",
			Path:   "/areas/2",
			Body:   "{\n  \"name\": \"areaNotExist\" \n}",
			Status: status.AreaNotExist,
		},
		{
			Method: "PUT",
			Path:   "/areas/1",
			Body:   "{\n  \"name\": \"areaExist\" \n}",
			Status: 0,
		},
		// 前端传入空参
		{
			Method: "PUT",
			Path:   "/areas/1",
			Body:   "{\n  \"name\": \"\" \n}",
			Status: status.AreaNameInputNilErr,
		},
		// 前端传入参数长度超出限制
		{
			Method: "PUT",
			Path:   "/areas/1",
			Body:   "{\n  \"name\": \"ashdoiauhfuioasodiugfiuagoiuahoiudgfyausgtdoyftguyatgfydsoaihdushdiuf\" \n}",
			Status: status.AreaNameLengthLimit,
		},
		{
			Method: "DELETE",
			Path:   "/areas/1",
			Status: status.Deny,
		},
		{
			Method: "GET",
			Path:   "/areas/2",
			Status: status.AreaNotExist,
		},
		{
			Method: "GET",
			Path:   "/areas/1",
			Status: 0,
		},
		{
			Method: "DELETE",
			Path:   "/areas/1/users/1",
			Status: 0,
		},
		// 用户已退出当前area
		{
			Method: "DELETE",
			Path:   "/areas/1/users/2",
			Status: status.RequireLogin,
		},
	}

	//先添加一条area记录
	test.CreateRecord(&entity.Area{Name: "demo"})

	//管理员角色用户的测试
	test.RunApiTest(t, RegisterAreaRouter, userCases, test.WithRoles("管理员"))

	adminCases := []test.ApiTestCase{
		{
			Method: "GET",
			Path:   "/areas",
			Status: 0,
			IsArray: []string{
				"data.areas",
			},
		},
		// 成员角色用户没有权限修改area
		{
			Method: "PUT",
			Path:   "/areas/1",
			Body:   "{\n  \"name\": \"area\" \n}",
			Status: status.Deny,
		},
		{
			Method: "DELETE",
			Path:   "/areas/1",
			Status: status.Deny,
		},
		{
			Method: "GET",
			Path:   "/areas/1",
			Status: 0,
		},
		{
			Method: "DELETE",
			Path:   "/areas/1/users/1",
			Status: 0,
		},
	}

	// 插入第二条记录
	test.CreateRecord(&entity.Area{Name: "demo"})
	// 成员角色用户的测试
	test.RunApiTest(t, RegisterAreaRouter, adminCases, test.WithRoles("成员"))
}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
