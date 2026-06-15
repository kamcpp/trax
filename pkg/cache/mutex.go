package cache

import (
	"context"
)

var defaultCache Cache

func SetDefaultCache(cache Cache) {
	defaultCache = cache
}

func Mutex(ctx context.Context, key string, ttlSec int, timeoutSec int64, cb func()) error {
	return defaultCache.Mutex(ctx, key, ttlSec, timeoutSec, cb)
}

// MultiMutex acquires multiple locks in a deterministic order (sorted keys) to prevent deadlocks.
// All locks must be acquired before the callback is executed.
// If any lock acquisition times out, all previously acquired locks are released.
func MultiMutex(ctx context.Context, keys []string, ttlSec int, timeoutSec int64, cb func()) error {
	return defaultCache.MultiMutex(ctx, keys, ttlSec, timeoutSec, cb)
}

// AcquireMultiLock acquires multiple distributed locks in a deterministic order (sorted keys).
// Unlike MultiMutex, this does NOT automatically release locks - caller must call ReleaseMultiLock.
// The TTL acts as a safety net: if the caller crashes or forgets to release, locks auto-expire.
// Use this for saga-level locks that span multiple async operations/steps.
// Returns nil on success, error if locks couldn't be acquired within timeoutSec.
func AcquireMultiLock(ctx context.Context, keys []string, ttlSec int, timeoutSec int64) error {
	return defaultCache.AcquireMultiLock(ctx, keys, ttlSec, timeoutSec)
}

// ReleaseMultiLock releases multiple previously acquired locks.
// Safe to call even if some locks have already expired or were never acquired.
func ReleaseMultiLock(ctx context.Context, keys []string) error {
	return defaultCache.ReleaseMultiLock(ctx, keys)
}
