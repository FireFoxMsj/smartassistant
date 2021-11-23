package generate

import (
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/oauth2.v3"
)

// NewAuthorizeGenerate create to generate the authorize code instance
func NewAuthorizeGenerate() *AuthorizeGenerate {
	return &AuthorizeGenerate{}
}

// AuthorizeGenerate generate the authorize code
type AuthorizeGenerate struct{}

// Token based on the UUID generated token
func (ag *AuthorizeGenerate) Token(data *oauth2.GenerateBasic) (string, error) {
	jwtGeneRate := NewJWTAccessGenerate(jwt.SigningMethodHS256)

	access, _, err := jwtGeneRate.Token(data, false)
	if err != nil {
		return "", err
	}
	return access, nil
}
