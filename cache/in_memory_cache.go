package cache

import "context"

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

func (cache *tinydatesInMemoryCache) Authorized(
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
