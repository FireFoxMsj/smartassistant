package store

import (
	"time"
)

const GoCacheType = "go-cache"

// GoCacheClientInterface represents a github.com/patrickmn/go-cache client
type GoCacheClientInterface interface {
	Get(k string) (interface{}, bool)
	GetWithExpiration(k string) (interface{}, time.Time, bool)
	Set(k string, x interface{}, d time.Duration)
	Delete(k string)
	Flush()
}

type GoCacheStore struct {
	client  GoCacheClientInterface
	options *Options
}

func NewGoCache(client GoCacheClientInterface, options *Options) *GoCacheStore {
	if options == nil {
		options = &Options{}
	}

	return &GoCacheStore{
		client:  client,
		options: options,
	}
}

func (s *GoCacheStore) Get(key interface{}) (interface{}, error) {
	var err error
	keyStr := key.(string)
	value, exists := s.client.Get(keyStr)
	if !exists {
		err = ErrValueNotFound
	}

	return value, err
}

func (s *GoCacheStore) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	data, t, exists := s.client.GetWithExpiration(key.(string))
	if !exists {
		return data, 0, ErrValueNotFound
	}
	duration := t.Sub(time.Now())
	return data, duration, nil
}

func (s *GoCacheStore) Set(key interface{}, value interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = s.options.Expiration
	}
	s.client.Set(key.(string), value, expiration)
	return nil
}

func (s *GoCacheStore) Delete(key interface{}) error {
	s.client.Delete(key.(string))
	return nil
}

func (s *GoCacheStore) Clear() error {
	s.client.Flush()
	return nil
}

func (g *GoCacheStore) GetType() string {
	return GoCacheType
}
