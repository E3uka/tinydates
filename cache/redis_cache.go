package cache

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

const SESSION_KEY = "tinysession"

// tinydatesRedisCache provide access to the Cache methods for a Redis backed
// cache.
type tinydatesRedisCache struct {
	Cache *redis.Pool
}

func NewTinydatesRedisCache(cache *redis.Pool) Cache {
	return &tinydatesRedisCache{Cache: cache}
}

func (cache *tinydatesRedisCache) StartSession(
	ctx context.Context,
	token string,
) error {
	conn := cache.Cache.Get()
	defer conn.Close()

	if _, err := conn.Do("SADD", SESSION_KEY, token); err != nil {
		return err
	}

	return nil
}

func (cache *tinydatesRedisCache) Authorized(
	ctx context.Context,
	token string,
) bool {
	conn := cache.Cache.Get()
	defer conn.Close()

	// for simplicity not handling errors and just returning key not found
	exists, err := redis.Bool(conn.Do("SISMEMBER", SESSION_KEY, token))
	if err != nil {
		return false
	}

	return exists

}

func (cache *tinydatesRedisCache) EndSession(
	ctx context.Context,
	token string,
) error {
	conn := cache.Cache.Get()
	defer conn.Close()

	if _, err := conn.Do("SREM", SESSION_KEY, token); err != nil {
		return err
	}

	return nil
}
