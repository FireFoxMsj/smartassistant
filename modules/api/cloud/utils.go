package cloud

import (
	"errors"
	"time"

	jwt2 "github.com/zhiting-tech/smartassistant/modules/utils/jwt"

	"github.com/dgrijalva/jwt-go"
	"github.com/zhiting-tech/smartassistant/modules/config"
)

var (
	ErrSAKeyEmpty = errors.New("sa key is empty")
)

type MigrationClaims struct {
	SAID string `json:"sa_id,omitempty"`
	Exp  int64  `json:"exp,omitempty"`
}

func (c MigrationClaims) Valid() error {
	if time.Unix(c.Exp, 0).Before(time.Now()) {
		return jwt2.ErrTokenIsExpired
	}
	return nil
}

func GenerateMigrationJwt(claims MigrationClaims) (jwtToken string, err error) {
	key := config.GetConf().SmartAssistant.Key
	if key == "" {
		return "", ErrSAKeyEmpty
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}
