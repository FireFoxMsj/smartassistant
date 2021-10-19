package cloud

import (
	"fmt"
	setting2 "github.com/zhiting-tech/smartassistant/modules/api/setting"
	"net/http"
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// bindCloudReq 绑定云端接口请求参数
type bindCloudReq struct {
	CloudAreaID string `json:"cloud_area_id"`
	CloudUserID int    `json:"cloud_user_id"`
}

type bindCloudResp struct {
	AreaID string `json:"area_id"` // 该AreaID用于客户端更新自己SC的家庭ID数据
}

// bindCloud 用于处理绑定云端接口的请求
func bindCloud(c *gin.Context) {

	var (
		req  bindCloudReq
		resp bindCloudResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	// 建立长连接
	saID := config.GetConf().SmartAssistant.ID
	scUrl := config.GetConf().SmartCloud.URL()
	cloudAreaID, err := strconv.ParseUint(req.CloudAreaID, 10, 64)
	if err != nil {
		err = errors.New(errors.BadRequest)
		return
	}
	// 更新用户和家庭关系
	url := fmt.Sprintf("%s/sa/%s/users/%d", scUrl, saID, req.CloudUserID)
	u := session.Get(c)
	saDevice, _ := entity.GetSaDevice()
	body := map[string]interface{}{
		"area_id":        cloudAreaID,
		"sa_user_id":     u.UserID,
		"sa_lan_address": saDevice.Address,
		"sa_area_id":     u.AreaID,
	}

	setting := entity.GetDefaultUserCredentialFoundSetting()
	if err = entity.GetSetting(entity.UserCredentialFoundType, &setting, u.AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 判断是否允许找回找回凭证
	if setting.UserCredentialFound {
		body["area_token"] = setting2.GetUserCredentialAuthToken()
	}

	_, err = cloud.CloudRequest(url, http.MethodPost, body)
	if err != nil {
		err = errors.New(status.SABindError)
		return
	}

	resp.AreaID = strconv.FormatUint(u.AreaID, 10)

}
