package cloud

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// CloudRequest 请求Cloud SC的方法
func CloudRequest(url, method string, requestData map[string]interface{}) (resp []byte, err error) {
	content, _ := json.Marshal(&requestData)

	logger.Println(url)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(content))
	if err != nil {
		logger.Error("new request error:", err.Error())
		return
	}

	req.Header = GetCloudReqHeader()
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("do request error:", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		logger.Println("http status: %v", response.Status)
		return resp, http.ErrNotSupported
	}

	defer response.Body.Close()
	resp, _ = ioutil.ReadAll(response.Body)
	return
}

// GetCloudReqHeader 获取请求Cloud SC的Header
func GetCloudReqHeader() http.Header {
	saID := config.GetConf().SmartAssistant.ID
	saKey := config.GetConf().SmartAssistant.Key
	header := http.Header{}
	header.Set(types.SAID, saID)
	header.Set(types.SAKey, saKey)
	return header
}
