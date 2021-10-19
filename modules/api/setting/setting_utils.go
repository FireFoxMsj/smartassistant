package setting

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	jwt2 "github.com/zhiting-tech/smartassistant/modules/utils/jwt"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

const (
	HttpRequestTimeout = (time.Duration(30) * time.Second)
)

var (
	once      sync.Once
	authToken string
)

type authTokenClaims struct {
	SAID string `json:"sa_id,omitempty"`
	Exp  int64  `json:"exp,omitempty"`
}

func (c authTokenClaims) Valid() error {
	if time.Unix(c.Exp, 0).Before(time.Now()) {
		return jwt2.ErrTokenIsExpired
	}
	return nil
}

// GenerateAuthTokenJwt 以 SAID 作为加密串生成 JWT
func GenerateAuthTokenJwt(claims authTokenClaims) (jwtToken string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(claims.SAID))
}

// validateAuthTokenJwt 校验找回用户凭证 JWT 的有效性
func ValidateAuthTokenJwt(jwtToken string) (*authTokenClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &authTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		claims, ok := token.Claims.(*authTokenClaims)
		if !ok {
			return nil, jwt2.ErrTokenNotValid
		}

		if claims.SAID != config.GetConf().SmartAssistant.ID {
			return nil, jwt2.ErrTokenNotValid
		}

		return []byte(config.GetConf().SmartAssistant.ID), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*authTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, jwt2.ErrTokenNotValid
	}
}

// TODO: 等修改用户的token字段后,需要修改为scope_token的实现方式
func GetUserCredentialAuthToken() string {
	once.Do(func() {
		token, _ := GenerateAuthTokenJwt(authTokenClaims{
			SAID: config.GetConf().SmartAssistant.ID,
			Exp:  time.Now().Add(876000 * time.Hour).Unix(), // 设置长的时间过期时间当作永不过期
		})
		authToken = token
	})

	return authToken
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
		"area_token": GetUserCredentialAuthToken(),
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

// TODO 找回用户凭证整合成oauth2模式
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
