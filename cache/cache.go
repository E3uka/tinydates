package cache

import "context"

// Cache are the core methods required from the cache for tinydates.
type Cache interface {
	// StartSession inserts a new token into the cache
	StartSession(ctx context.Context, token string) error

	// Exists removes a token from the cache
	Authenticated(ctx context.Context, token string) bool

	// EndSession removes a token from the cache
	EndSession(ctx context.Context, token string) error
}
