package store

import (
	"time"
)

type StoreInterface interface {
	Get(key interface{}) (interface{}, error)
	GetWithTTL(key interface{}) (interface{}, time.Duration, error)
	Set(key interface{}, value interface{}, expir time.Duration) error
	Delete(key interface{}) error
	Clear() error
	GetType() string
}
