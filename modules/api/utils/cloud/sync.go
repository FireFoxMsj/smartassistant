package cloud

import (
	"fmt"
	"net/http"

	"github.com/zhiting-tech/smartassistant/modules/config"
)

func RemoveSA(areaID uint64) {
	request := map[string]interface{}{"is_remove": true}
	syncToCloud(areaID, request)
}
func RemoveSAUser(areaID uint64, userID int) {
	request := map[string]interface{}{"remove_sa_user_id": userID}
	syncToCloud(areaID, request)
}
func UpdateAreaName(areaID uint64, name string) {
	request := map[string]interface{}{"sa_area_name": name}
	syncToCloud(areaID, request)
}
func syncToCloud(areaID uint64, request map[string]interface{}) {
	// 建立长连接
	scUrl := config.GetConf().SmartCloud.URL()
	saID := config.GetConf().SmartAssistant.ID
	// 更新用户和家庭关系
	url := fmt.Sprintf("%s/sa/%s/areas/%d", scUrl, saID, areaID)

	_, err := CloudRequest(url, http.MethodPut, request)
	if err != nil {
		return
	}
}
