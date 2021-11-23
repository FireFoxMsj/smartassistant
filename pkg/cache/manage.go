package cache

import (
	"github.com/patrickmn/go-cache"
	"github.com/zhiting-tech/smartassistant/pkg/cache/store"
	"time"
)

var (
	cacheManager = newDefault()
)

// 使用自定义的存储初始化全局cache。
func InitCache(store store.StoreInterface) {
	cacheManager = New(store)
}

// 使用go-cache创建默认的缓存。
func newDefault() *Cache {
	gocacheClient := cache.New(5*time.Minute, 10*time.Minute)
	gocacheStore := store.NewGoCache(gocacheClient, nil)
	return New(gocacheStore)
}

// 根据key获取value，如果获取不到，返回ErrValueNotFound错误。
func Get(key interface{}) (interface{}, error) {
	return cacheManager.Get(key)
}

// 根据key获取value和value在缓存的有限期。
func GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	return cacheManager.GetWithTTL(key)
}

// 添加key对应的value到缓存中，有效期expiration为0，
// 则使用创建缓存存储时的options的expiration属性，默认缓存存储go-cache的过期时间为5分钟。
func Set(key interface{}, val interface{}, expiration time.Duration) error {
	return cacheManager.Set(key, val, expiration)
}

// 删除key对应的缓存。
func Delete(key interface{}) error {
	return cacheManager.Delete(key)
}

// 获取缓存的存储。
func GetStore() store.StoreInterface {
	return cacheManager.GetStore()
}

// 获取缓存的存储类型。
func GetType() string {
	return cacheManager.GetType()
}
