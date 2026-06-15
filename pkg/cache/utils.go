package cache

import (
	"context"
)

func CacheRead(
	ctx context.Context,
	db int,
	key string,
	ttlSec uint,
	missCb func() (string, error),
) (string, error) {
	return defaultCache.CacheRead(ctx, db, key, ttlSec, missCb)
}

func CacheRead2(
	ctx context.Context,
	key string,
	missCb func() (string, error),
) (string, error) {
	return defaultCache.CacheRead2(ctx, key, missCb)
}

func CacheRead3(
	ctx context.Context,
	db int,
	key string,
	ttlSec uint,
	missCb func() (string, error),
	hitCb func(string) (bool, string, bool, error),
) (string, error) {
	return defaultCache.CacheRead3(ctx, db, key, ttlSec, missCb, hitCb)
}
