package cache

import (
	cache2 "github.com/patrickmn/go-cache"
	"sync"
	"time"
)

var c *cache2.Cache
var cacheOnce sync.Once

func GetCache() *cache2.Cache {
	cacheOnce.Do(func() {
		c = cache2.New(5*time.Minute, 10*time.Minute)
	})

	return c
}

// GetValWithCode 通过code获取对应的值
func GetValWithCode(code string) string {
	ca := GetCache()
	v, isExist := ca.Get(code)
	if !isExist {
		return ""
	}

	// 验证成功删除该验证码
	val, ok := v.(string)
	if !ok {
		return ""
	}
	return val
}
