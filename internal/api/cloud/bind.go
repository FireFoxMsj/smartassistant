package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/config"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// bindCloudReq 绑定云端接口请求参数
type bindCloudReq struct {
	CloudToken  string `json:"cloud_token"` // TODO 云端颁发的token
	CloudAreaID int    `json:"cloud_area_id"`
	CloudUserID int    `json:"cloud_user_id"`
}

// bindCloud 用于处理绑定云端接口的请求
func bindCloud(c *gin.Context) {

	var (
		req  bindCloudReq
		resp interface{}
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	// 建立长连接
	saID := config.GetConf().SmartAssistant.ID
	scUrl := config.GetConf().SmartCloud.URL()
	// 更新用户和家庭关系
	url := fmt.Sprintf("%s/sa/%s/users/%d", scUrl, saID, req.CloudUserID)
	u := session.Get(c)
	saDevice, _ := entity.GetSaDevice()
	body := map[string]interface{}{
		"area_id":        req.CloudAreaID,
		"sa_user_id":     u.UserID,
		"sa_user_token":  u.Token,
		"sa_id":          saID,
		"sa_lan_address": saDevice.Address,
	}
	b, _ := json.Marshal(body)
	log.Println(url)
	bindReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
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
