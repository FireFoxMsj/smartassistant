package store

import (
	"github.com/go-redis/redis"
	"time"
)

type RedisClientInterface interface {
	Get(key string) *redis.StringCmd
	TTL(key string) *redis.DurationCmd
	Set(key string, values interface{}, expiration time.Duration) *redis.StatusCmd
	Del(keys ...string) *redis.IntCmd
	FlushAll() *redis.StatusCmd
}

const RedisType = "redis"

type RedisStore struct {
	client  RedisClientInterface
	options *Options
}

func NewRedis(client RedisClientInterface, options *Options) *RedisStore {
	if options == nil {
		options = &Options{}
	}

	return &RedisStore{
		client:  client,
		options: options,
	}
}

func (s *RedisStore) Get(key interface{}) (interface{}, error) {
	object, err := s.client.Get(key.(string)).Result()
	if err != nil {
		err = ErrValueNotFound
	}
	return object, err
}

func (s *RedisStore) GetWithTTL(key interface{}) (interface{}, time.Duration, error) {
	object, err := s.client.Get(key.(string)).Result()
	if err != nil {
		return nil, 0, ErrValueNotFound
	}

	ttl, err := s.client.TTL(key.(string)).Result()
	if err != nil {
		return nil, 0, ErrValueNotFound
	}

	return object, ttl, err
}

func (s *RedisStore) Set(key interface{}, val interface{}, expiration time.Duration) error {
	if expiration == 0 {
		expiration = s.options.Expiration
	}

	return s.client.Set(key.(string), val, expiration).Err()
}

func (s *RedisStore) Delete(key interface{}) error {
	_, err := s.client.Del(key.(string)).Result()
	return err
}

func (s *RedisStore) Clear() error {
	return s.client.FlushAll().Err()
}

func (s *RedisStore) GetType() string {
	return RedisType
}
