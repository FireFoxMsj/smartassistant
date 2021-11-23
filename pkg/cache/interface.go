package cache

import "github.com/zhiting-tech/smartassistant/pkg/cache/store"

type CacheInterface interface {
	store.StoreInterface

	GetStore() store.StoreInterface
}
