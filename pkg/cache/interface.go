package cache

import "context"

type ZMember struct {
	Score  float64
	Member string
}

type Cache interface {
	Init(ctx context.Context) error

	Exists(ctx context.Context, db int, key string) (bool, error)
	Get(ctx context.Context, db int, key string) (string, error)
	Set(ctx context.Context, db int, key, value string, ttlSec uint) error
	Del(ctx context.Context, db int, key string) error

	ListPushLeft(ctx context.Context, db int, key, element string) error
	ListPopRight(ctx context.Context, db int, key string) (string, error)
	ListLen(ctx context.Context, db int, key string) (int64, error)
	ListRange(ctx context.Context, db int, key string, start, stop int64) ([]string, error)
	RPush(ctx context.Context, db int, key, element string) error

	AddToSet(ctx context.Context, db int, key, element string) error
	IsSetMember(ctx context.Context, db int, key, element string) (bool, error)
	SetMembers(ctx context.Context, db int, key string) ([]string, error)

	// Hash operations
	HSet(ctx context.Context, db int, key, field, value string) error
	HGet(ctx context.Context, db int, key, field string) (string, error)
	HGetAll(ctx context.Context, db int, key string) (map[string]string, error)
	HDel(ctx context.Context, db int, key string, fields ...string) error
	HMSet(ctx context.Context, db int, key string, fields map[string]string) error

	// Sorted set operations
	ZAdd(ctx context.Context, db int, key string, score float64, member string) error
	ZRem(ctx context.Context, db int, key string, members ...string) error
	ZRangeByScore(ctx context.Context, db int, key string, min, max string, offset, count int64) ([]string, error)
	ZRevRangeByScore(ctx context.Context, db int, key string, max, min string, offset, count int64) ([]string, error)
	ZCard(ctx context.Context, db int, key string) (int64, error)
	ZRangeWithScores(ctx context.Context, db int, key string, start, stop int64) ([]ZMember, error)
	ZRevRangeWithScores(ctx context.Context, db int, key string, start, stop int64) ([]ZMember, error)
	ZRemRangeByScore(ctx context.Context, db int, key, min, max string) error

	JsonSet(ctx context.Context, db int, key, path, value string) error
	JsonGet(ctx context.Context, db int, key, path string) (string, error)
	JsonArrAppend(ctx context.Context, db int, key, path, element string) error

	Mutex(ctx context.Context, key string, ttlSec int, timeoutSec int64, cb func()) error
	// MultiMutex acquires multiple locks in a deterministic order (sorted keys) to prevent deadlocks.
	// All locks must be acquired before the callback is executed.
	// If any lock acquisition times out, all previously acquired locks are released.
	MultiMutex(ctx context.Context, keys []string, ttlSec int, timeoutSec int64, cb func()) error

	// AcquireMultiLock acquires multiple distributed locks in a deterministic order (sorted keys).
	// Unlike MultiMutex, this does NOT automatically release locks - caller must call ReleaseMultiLock.
	// The TTL acts as a safety net: if the caller crashes or forgets to release, locks auto-expire.
	// Use this for saga-level locks that span multiple async operations/steps.
	// Returns nil on success, error if locks couldn't be acquired within timeoutSec.
	AcquireMultiLock(ctx context.Context, keys []string, ttlSec int, timeoutSec int64) error

	// ReleaseMultiLock releases multiple previously acquired locks.
	// Safe to call even if some locks have already expired or were never acquired.
	ReleaseMultiLock(ctx context.Context, keys []string) error

	CacheRead(ctx context.Context, db int, key string, ttlSec uint, missCb func() (string, error)) (string, error)
	CacheRead2(ctx context.Context, key string, missCb func() (string, error)) (string, error)
	CacheRead3(ctx context.Context, db int, key string, ttlSec uint, missCb func() (string, error), hitCb func(string) (bool, string, bool, error)) (string, error)

	Keys(ctx context.Context, db int, pattern string) ([]string, error)
}
