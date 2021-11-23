package clouddisk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/server"
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
	u := session.Get(c)
	accessToken := c.GetHeader(types.SATokenKey)
	ti, _ := oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
	c.Request.Header.Set(types.AreaID, strconv.FormatUint(u.AreaID, 10))
	tgr := &server.AuthorizeRequest{
		ResponseType:   oauth2.Token,
		ClientID:       ti.GetClientID(),
		UserID:         ti.GetUserID(),
		Scope:          strings.Join([]string{"area", "user"}, ","),
		AccessTokenExp: 10 * time.Minute,
		Request:        c.Request,
	}

	tokenInfo, err := oauth.GetOauthServer().GetAuthorizeToken(tgr)
	if err != nil {
		logger.Errorf("get token failed err: (%v)", err)
		return
	}

	// 4、获取用户id,scope-token并放入header
	request.Header.Set("scope-user-id", strconv.Itoa(u.UserID))
	request.Header.Set("scope-token", tokenInfo.GetAccess())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(request)
	if err != nil || resp.StatusCode != http.StatusOK {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}
