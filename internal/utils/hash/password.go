package hash

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"github.com/gorilla/securecookie"
	"strings"
)

const (
	saTokenName = "smart assistant"
	saSecretKey = "!!!secret_key_for_smart_assistant!!!"
)

func GenerateHashedPassword(password, salt string) string {
	tmp := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(tmp[:])
}

func CheckPassword(password, salt, hashedPassword string) bool {
	return GenerateHashedPassword(password, salt) == hashedPassword
}

// GetSaToken 生成sa的token
func GetSaToken() (saToken string) {
	tokenId := strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
	codeCs := securecookie.CodecsFromPairs([]byte(saSecretKey))
	saToken, _ = securecookie.EncodeMulti(saTokenName, tokenId, codeCs...)
	return
}
