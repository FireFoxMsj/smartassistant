package scene

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhiting-tech/smartassistant/modules/api/test"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"testing"
)

func TestScene(t *testing.T) {
	cases := []test.ApiTestCase{
		// 获取场景
		{
			Method:  "GET",
			Path:    "/scenes",
			Status:  0,
			IsArray: []string{"data.manual"},
		},
		// 获取场景详情
		{
			Method: "GET",
			Path:   "/scenes/1",
			Status: 0,
			IsID:   []string{"data.id"},
		},
		// 不属于用户的场景详情
		{
			Method: "GET",
			Path:   "/scenes/2",
			Status: status.Deny,
		},
		// 场景的执行
		{
			Method: "POST",
			Path:   "/scenes/1/execute",
			Body:   "{\"is_execute\": true}",
			Status: 0,
		},
		// 不属于用户的场景的执行
		{
			Method: "POST",
			Path:   "/scenes/2/execute",
			Body:   "{\"is_execute\": true}",
			Status: status.Deny,
		},
		// 创建场景
		{
			Method: "POST",
			Path:   "/scenes",
			Body:   "{  \"name\": \"abc\",  \"auto_run\": false,  \"time_period\": 1,  \"repeat_date\": \"1\",  \"scene_tasks\": [{ \"type\": 2, \"delay_seconds\": 2, \"control_scene_id\": 1} ]}",
			Status: 0,
		},
		// 修改场景
		{
			Method: "PUT",
			Path:   "/scenes/1",
			Body:   "{  \"id\": 1, \"name\": \"demo_changed\", \"time_period\": 1,  \"repeat_date\": \"1\",  \"scene_tasks\": \n[\n    { \n        \"id\": 1,\n        \"scene_id\": 1,\n        \"type\": 1, \n        \"delay_seconds\": 9, \n        \"control_scene_id\": 2,\n        \"type\": 2\n    } \n]\n}",
			Status: 0,
		},
		// 删除场景
		{
			Method: "DELETE",
			Path:   "/scenes/1",
			Status: 0,
		},
		// 删除不属于用户的场景
		{
			Method: "DELETE",
			Path:   "/scenes/2",
			Status: status.Deny,
		},
	}

	// 创建测试场景
	test.CreateRecord(&entity.Scene{Name: "demo", AreaID: 1})
	test.CreateRecord(&entity.Scene{Name: "demo_delete", AreaID: 2})
	test.CreateRecord(&entity.Device{Name: "demo", Model: "smart_assistant", OwnerID: 1})

	test.RunApiTest(t, InitSceneRouter, cases, test.WithRoles("管理员"), test.WithAreas(1))

	var s1, s2 entity.Scene
	// 验证场景创建
	db := entity.GetDB().Where("name=?", "abc").Find(&s1)
	assert.Empty(t, db.Error)
	assert.NotEmpty(t, s1.ID, "abc")
	// 验证场景修改
	db = entity.GetDB().Where("name=?", "demo").Find(&s2)
	assert.Empty(t, db.Error)
	assert.Empty(t, s2.ID, "demo")
}

func TestMain(m *testing.M) {
	test.InitApiTest(m)
}
