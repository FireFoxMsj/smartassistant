package clouddisk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	jwt2 "github.com/zhiting-tech/smartassistant/modules/utils/jwt"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DelAreaCloudDisk 删除家庭的网盘资源
func DelAreaCloudDisk(c *gin.Context, areaID uint64) (err error) {

	ids, err := entity.GetUIds(areaID)
	if err != nil {
		return
	}

	if len(ids) == 0 {
		return
	}
	if err = DelCloudDisk(c, ids...); err != nil {
		return
	}

	return
}

// DelCloudDisk 删除网盘资源
func DelCloudDisk(c *gin.Context, ids ...int) (err error) {
	url := fmt.Sprintf("http://%s/api/plugin/wangpan/folders", types.CloudDiskAddr)

	param := map[string]interface{}{
		"user_ids": ids,
	}

	// 2  序列化数据
	content, _ := json.Marshal(param)

	request, err := http.NewRequest("DELETE", url, bytes.NewReader(content))
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 3 获取scope_token
	claims := jwt2.AccessClaims{
		UID:   session.Get(c).UserID,
		SAID:  config.GetConf().SmartAssistant.ID,
		Exp:   time.Now().Add(10 * time.Minute).Unix(),
		Scope: strings.Join([]string{"area", "user"}, ","),
	}

	token, err := jwt2.GenerateUserJwt(claims, session.Get(c))
	if err != nil {
		log.Printf("generate jwt error %s", err.Error())
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	// 4、获取用户id,scope-token并放入header
	request.Header.Set("scope-user-id", strconv.Itoa(session.Get(c).UserID))
	request.Header.Set("scope-token", token)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(request)
	if err != nil || resp.StatusCode != http.StatusOK {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}
