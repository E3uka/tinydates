package cache

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

// Cache are the core methods required from the cache for tinydates.
type Cache interface {
	// StartSession inserts a new token into the cache
	StartSession(ctx context.Context, token string) error

	// Exists removes a token from the cache
	Authenticated(ctx context.Context, token string) bool

	// EndSession removes a token from the cache
	EndSession(ctx context.Context, token string) error
}

const SESSION_KEY = ""

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

	if err := conn.Send("SADD", SESSION_KEY, token); err != nil {
		return err
	}
	if err := conn.Flush(); err != nil {
		return err
	}

	return nil
}

func (cache *tinydatesRedisCache) Authenticated(
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

	if err := conn.Send("SREM", SESSION_KEY, token); err != nil {
		return err
	}
	if err := conn.Flush(); err != nil {
		return err
	}

	return nil
}

// tinydatesInMemoryCache provide access to the Cache methods for an in memory
// cache, this uses a map of string to interface that effectively works as a
// set allowing only for unique values as the key, the `interface{}` value
// takes up so memory so is not stored - demonstrating new dependency injection
// during testing
type tinydatesInMemoryCache struct {
	Cache map[string]interface{}
}

func NewTinydatesInMemoryCache() Cache {
	return &tinydatesInMemoryCache{Cache: make(map[string]interface{})}
}

func (cache *tinydatesInMemoryCache) StartSession(
	ctx context.Context, 
	token string,
) error {
	cache.Cache[token] = struct{}{}
	return nil
}

func (cache *tinydatesInMemoryCache) Authenticated(
	ctx context.Context, 
	token string,
) bool {
	_, exists := cache.Cache[token]
	return exists
}

func (cache *tinydatesInMemoryCache) EndSession(
	ctx context.Context, 
	token string,
) error {
	delete(cache.Cache, token)
	return nil
}

type TestCache interface {
	Cache
}
