package cache

import (
	"crypto"
	"fmt"
	"github.com/zhiting-tech/smartassistant/pkg/cache/store"
	"reflect"
	"time"
)

type Cache struct {
	store store.StoreInterface
}

func New(store store.StoreInterface) *Cache {
	return &Cache{store: store}
}

func (c *Cache) getCacheKey(key interface{}) string {
	switch key.(type) {
	case string:
		return key.(string)
	default:
		return checksum(key)
	}
}

func checksum(object interface{}) string {
	digester := crypto.MD5.New()
	fmt.Fprint(digester, reflect.TypeOf(object))
	fmt.Fprint(digester, object)
	hash := digester.Sum(nil)

	return fmt.Sprintf("%x", hash)
}

func (c *Cache) Get(key interface{}) (interface{}, error) {
	cacheKey := c.getCacheKey(key)
	return c.store.Get(cacheKey)
}

func (c *Cache) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	cacheKey := c.getCacheKey(key)
	return c.store.GetWithTTL(cacheKey)
}

func (c *Cache) Set(key interface{}, val interface{}, expiration time.Duration) error {
	cacheKey := c.getCacheKey(key)
	return c.store.Set(cacheKey, val, expiration)
}

func (c *Cache) Delete(key interface{}) error {
	cacheKey := c.getCacheKey(key)
	return c.store.Delete(cacheKey)
}

func (c *Cache) Clear() error {
	return c.store.Clear()
}

func (c *Cache) GetType() string {
	return c.store.GetType()
}

func (c *Cache) GetStore() store.StoreInterface {
	return c.store
}
