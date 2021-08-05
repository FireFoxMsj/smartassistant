package jwt

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	orm2 "github.com/zhiting-tech/smartassistant/internal/entity"
	session2 "github.com/zhiting-tech/smartassistant/internal/utils/session"
)

var (
	ErrEmptySignedKey = errors.New("jwt signed key is empty")
	ErrTokenNotValid  = errors.New("jwt token is not valid")
)

// GenerateUserJwt 以用户 token 作为加密串生成 JWT，用户 ID 使用 uid 字段
func GenerateUserJwt(claims jwt.MapClaims, u *session2.User) (jwtToken string, err error) {
	if u == nil || len(u.Token) == 0 {
		return "", ErrEmptySignedKey
	}
	if _, ok := claims["uid"]; !ok {
		claims["uid"] = u.UserID
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.Token))
}

// ValidateUserJwt 校验用户 JWT 的有效性
func ValidateUserJwt(jwtToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, ErrTokenNotValid
		}
		userId, ok := claims["uid"].(float64)
		if !ok {
			return nil, ErrTokenNotValid
		}
		user, err := orm2.GetUserByID(int(userId))
		if err != nil {
			return nil, err
		}
		return []byte(user.Token), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, ErrTokenNotValid
	}
}
