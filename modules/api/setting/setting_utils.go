package setting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"gopkg.in/oauth2.v3"
	"net/http"
	"strconv"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/entity"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

const (
	HttpRequestTimeout = (time.Duration(30) * time.Second)
)

func GetUserCredentialAuthToken(areaID uint64) string {
	req, _ := http.NewRequest("", "", nil)
	req.Header.Set(types.AreaID, strconv.FormatUint(areaID, 10))
	scClient, _ := entity.GetSCClient()
	tgr := oauth2.TokenGenerateRequest{
		ClientID:     scClient.ClientID,
		ClientSecret: scClient.ClientSecret,
		Scope:        scClient.AllowScope,
		Request:      req,
	}

	ti, err := oauth.GetOauthServer().GetAccessToken(oauth2.ClientCredentials, &tgr)
	if err != nil {
		logger.Errorf("get access token failed: (%v)", err)
		return ""
	}

	return ti.GetAccess()
}

// SendUserCredentialAuthTokenToSC 发送找回用户凭证的认证token给SC
func SendUserCredentialAuthTokenToSC(areaID uint64) {
	if len(config.GetConf().SmartCloud.Domain) <= 0 {
		return
	}
	saID := config.GetConf().SmartAssistant.ID
	scUrl := config.GetConf().SmartCloud.URL()
	url := fmt.Sprintf("%s/sa/%s/areas/%d", scUrl, saID, areaID)
	body := map[string]interface{}{
		"area_token": GetUserCredentialAuthToken(areaID),
	}
	b, _ := json.Marshal(body)
	logger.Debug(url)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
	if err != nil {
		logger.Warnf("NewRequest error %v\n", err)
		return
	}

	req.Header = cloud.GetCloudReqHeader()
	ctx, _ := context.WithTimeout(context.Background(), HttpRequestTimeout)
	req.WithContext(ctx)
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Warnf("request %s error %v\n", url, err)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		logger.Warnf("request %s error,status:%v\n", url, httpResp.Status)
		return
	}
}

func SendUserCredentialToSC() {
	areas, err := entity.GetAreas()
	if err != nil {
		logger.Errorf("get areas err (%v)", err)
		return
	}

	for _, area := range areas {
		SendUserCredentialAuthTokenToSC(area.ID)
	}

}
