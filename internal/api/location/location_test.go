package location

import (
	"github.com/zhiting-tech/smartassistant/internal/api/test"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"testing"
)

// TestLocation 对location API进行测试
func TestLocation(t *testing.T) {
	cases := []test.ApiTestCase{
		{
			Method: "GET",
			Path:   "/location_tmpl",
			Status: 0,
			Reason: "成功",
			IsArray: []string{
				"data.locations",
			},
		},
		{
			Method: "POST",
			Path:   "/locations",
			Body:   "{\n  \"name\": \"浴室\" \n}",
			Status: 0,
			Reason: "成功",
		},
		// 前端传入空参
		{
			Method: "POST",
			Path:   "/locations",
			Body:   "{\n  \"name\": \"\" \n}",
			Status: status.LocationNameInputNilErr,
		},
		// 前端传入参数长度超出限制
		{
			Method: "POST",
			Path:   "/locations",
			Body:   "{\n  \"name\": \"sadhaisuhduiahysiugsadyufstgdisuydfiusydiuayisduyaiusdydas\" \n}",
			Status: status.LocationNameLengthLimit,
		},
		{
			Method: "PUT",
			Path:   "/locations/1",
			Body:   "{\n  \"name\": \"厨房\"}",
			Status: 0,
		},
		{
			Method: "DELETE",
			Path:   "/locations/1",
			Status: 0,
		},
		// 测试过程中间插入一天记录
		{
			Method: "POST",
			Path:   "/locations",
			Body:   "{\n  \"name\": \"浴室\" \n}",
			Status: 0,
		},
		{
			Method: "GET",
			Path:   "/locations/1",
			Status: 0,
			IsArray: []string{
				"data.devices",
			},
		},
		{
			Method: "GET",
			Path:   "/locations/1/devices",
			Status: 0,
			IsArray: []string{
				"data.devices",
			},
		},
		{
			Method: "GET",
			Path:   "/locations",
			Status: 0,
			IsArray: []string{
				"data.locations",
			},
			IsID: []string{
				"data.locations.0.id",
			},
		},
		{
			Method: "PUT",
			Path:   "/locations",
			Body:   "{\n  \"locations_id\": [1] \n}",
			Status: 0,
		},
	}

	test.RunApiTest(t, RegisterLocationRouter, cases, test.WithRoles("管理员"))

}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
