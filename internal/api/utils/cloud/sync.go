package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zhiting-tech/smartassistant/internal/config"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"log"
	"net/http"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

func RemoveSA() {
	request := map[string]interface{}{"is_remove": true}
	syncToCloud(request, "") // TODO 后续补充token
}
func RemoveSAUser(userID int) {
	request := map[string]interface{}{"remove_sa_user_id": userID}
	syncToCloud(request, "")
}
func UpdateAreaName(name string) {
	request := map[string]interface{}{"sa_area_name": name}
	syncToCloud(request, "")
}
func syncToCloud(request map[string]interface{}, token string) {
	request["token"] = token

	// 建立长连接
	saID := config.GetConf().SmartAssistant.ID
	scUrl := config.GetConf().SmartCloud.URL()
	// 更新用户和家庭关系
	url := fmt.Sprintf("%s/sa/%s", scUrl, saID)
	b, _ := json.Marshal(request)
	log.Println(url)
	bindReq, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
	if err != nil {
		return
	}
	httpResp, err := http.DefaultClient.Do(bindReq)
	if err != nil {
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		log.Println("request error,status:", httpResp.Status)
		err = errors.New(status.SABindError)
		return
	}
}
