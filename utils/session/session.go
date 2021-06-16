package session

import (
	"crypto/sha256"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"sync"
)

const (
	DefaultSessionName = "_session_"
)

var (
	once        sync.Once
	cookieStore cookie.Store
)

func initStore() {
	h := sha256.New()
	// TODO secret key 从配置文件获取
	h.Write([]byte("!!!secret_key_for_test!!!"))
	res := h.Sum(nil)
	cookieStore = cookie.NewStore(res[:16], res[16:])
	cookieStore.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 30,
	})
}

func GetStore() sessions.Store {
	once.Do(initStore)
	return cookieStore
}

func GetSession(ctx *gin.Context) sessions.Session {
	return sessions.Default(ctx)
}
