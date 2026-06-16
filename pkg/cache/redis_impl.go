package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/xshyft/trax/pkg/common"
)

var (
	RedisAddr string
)

type RedisCache struct {
	rdb []*redis.Client
}

func NewRedisCache() *RedisCache {
	return &RedisCache{}
}

func (r *RedisCache) Init(ctx context.Context) error {
	if len(RedisAddr) == 0 {
		RedisAddr = os.Getenv("REDIS_ADDRESS")
		if len(RedisAddr) == 0 {
			panic("REDIS_ADDRESS is not set")
		}
	}
	r.rdb = make([]*redis.Client, 16)
	for i := 0; i < len(r.rdb); i++ {
		r.rdb[i] = redis.NewClient(&redis.Options{
			Addr:     RedisAddr,
			Password: "",
			DB:       i,
		})
	}
	for i := 0; i < len(r.rdb); i++ {
		err := r.Set(ctx, i, "test-key", "test-value", 300)
		if err != nil {
			return err
		}
	}
	common.L.Info("[cache] redis connections are initialized", common.F(ctx)...)
	return nil
}

func (r *RedisCache) Exists(ctx context.Context, db int, key string) (bool, error) {
	val, err := r.rdb[db].Exists(ctx, key).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return false, err
	}
	return val == 1, nil
}

func (r *RedisCache) Get(ctx context.Context, db int, key string) (string, error) {
	val, err := r.rdb[db].Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return "", err
	}
	return val, nil
}

func (r *RedisCache) Set(ctx context.Context, db int, key, value string, ttlSec uint) error {
	_, err := r.rdb[db].Set(ctx, key, value, time.Duration(ttlSec)*time.Second).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

func (r *RedisCache) Del(ctx context.Context, db int, key string) error {
	_, err := r.rdb[db].Del(ctx, key).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

func (r *RedisCache) ListPushLeft(ctx context.Context, db int, key, element string) error {
	_, err := r.rdb[db].LPush(ctx, key, element).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

func (r *RedisCache) ListPopRight(ctx context.Context, db int, key string) (string, error) {
	val, err := r.rdb[db].RPop(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return "", err
	}
	return val, nil
}

func (r *RedisCache) ListLen(ctx context.Context, db int, key string) (int64, error) {
	val, err := r.rdb[db].LLen(ctx, key).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return 0, err
	}
	return val, nil
}

func (r *RedisCache) ListRange(ctx context.Context, db int, key string, start, stop int64) ([]string, error) {
	val, err := r.rdb[db].LRange(ctx, key, start, stop).Result()
	if err == redis.Nil {
		return []string{}, nil
	} else if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return nil, err
	}
	return val, nil
}

func (r *RedisCache) AddToSet(ctx context.Context, db int, key, element string) error {
	_, err := r.rdb[db].SAdd(ctx, key, element).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

func (r *RedisCache) IsSetMember(ctx context.Context, db int, key, element string) (bool, error) {
	isMember, err := r.rdb[db].SIsMember(ctx, key, element).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return false, err
	}
	return isMember, nil
}

func (r *RedisCache) SetMembers(ctx context.Context, db int, key string) ([]string, error) {
	val, err := r.rdb[db].SMembers(ctx, key).Result()
	if err == redis.Nil {
		return []string{}, nil
	} else if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return nil, err
	}
	return val, nil
}

func (r *RedisCache) RPush(ctx context.Context, db int, key, element string) error {
	_, err := r.rdb[db].RPush(ctx, key, element).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

// Hash operations

func (r *RedisCache) HSet(ctx context.Context, db int, key, field, value string) error {
	_, err := r.rdb[db].HSet(ctx, key, field, value).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

func (r *RedisCache) HGet(ctx context.Context, db int, key, field string) (string, error) {
	val, err := r.rdb[db].HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return "", err
	}
	return val, nil
}

func (r *RedisCache) HGetAll(ctx context.Context, db int, key string) (map[string]string, error) {
	val, err := r.rdb[db].HGetAll(ctx, key).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return nil, err
	}
	return val, nil
}

func (r *RedisCache) HDel(ctx context.Context, db int, key string, fields ...string) error {
	_, err := r.rdb[db].HDel(ctx, key, fields...).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

func (r *RedisCache) HMSet(ctx context.Context, db int, key string, fields map[string]string) error {
	values := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		values = append(values, k, v)
	}
	_, err := r.rdb[db].HMSet(ctx, key, values...).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

// Sorted set operations

func (r *RedisCache) ZAdd(ctx context.Context, db int, key string, score float64, member string) error {
	_, err := r.rdb[db].ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

func (r *RedisCache) ZRem(ctx context.Context, db int, key string, members ...string) error {
	args := make([]interface{}, len(members))
	for i, m := range members {
		args[i] = m
	}
	_, err := r.rdb[db].ZRem(ctx, key, args...).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

func (r *RedisCache) ZRangeByScore(ctx context.Context, db int, key string, min, max string, offset, count int64) ([]string, error) {
	val, err := r.rdb[db].ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return nil, err
	}
	return val, nil
}

func (r *RedisCache) ZRevRangeByScore(ctx context.Context, db int, key string, max, min string, offset, count int64) ([]string, error) {
	val, err := r.rdb[db].ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return nil, err
	}
	return val, nil
}

func (r *RedisCache) ZCard(ctx context.Context, db int, key string) (int64, error) {
	val, err := r.rdb[db].ZCard(ctx, key).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return 0, err
	}
	return val, nil
}

func (r *RedisCache) ZRangeWithScores(ctx context.Context, db int, key string, start, stop int64) ([]ZMember, error) {
	val, err := r.rdb[db].ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return nil, err
	}
	result := make([]ZMember, len(val))
	for i, z := range val {
		result[i] = ZMember{Score: z.Score, Member: z.Member.(string)}
	}
	return result, nil
}

func (r *RedisCache) ZRevRangeWithScores(ctx context.Context, db int, key string, start, stop int64) ([]ZMember, error) {
	val, err := r.rdb[db].ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return nil, err
	}
	result := make([]ZMember, len(val))
	for i, z := range val {
		result[i] = ZMember{Score: z.Score, Member: z.Member.(string)}
	}
	return result, nil
}

func (r *RedisCache) ZRemRangeByScore(ctx context.Context, db int, key, min, max string) error {
	_, err := r.rdb[db].ZRemRangeByScore(ctx, key, min, max).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return err
	}
	return nil
}

func (r *RedisCache) JsonSet(ctx context.Context, db int, key, path, value string) error {
	result := r.rdb[db].Do(ctx, "JSON.SET", key, path, value)
	if result.Err() != nil {
		common.L.Warn(result.Err().Error(), common.F(ctx)...)
		return result.Err()
	}
	return nil
}

func (r *RedisCache) JsonGet(ctx context.Context, db int, key, path string) (string, error) {
	result := r.rdb[db].Do(ctx, "JSON.GET", key, path)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return "", nil
		} else {
			common.L.Warn(result.Err().Error(), common.F(ctx)...)
			return "", result.Err()
		}
	}
	return result.Text()
}

func (r *RedisCache) JsonArrAppend(ctx context.Context, db int, key, path, element string) error {
	result := r.rdb[db].Do(ctx, "JSON.ARRAPPEND", key, path, element)
	if result.Err() != nil {
		common.L.Warn(result.Err().Error(), common.F(ctx)...)
		return result.Err()
	}
	return nil
}

func (r *RedisCache) lockMutex(ctx context.Context, key string, ttlSec int) (bool, error) {
	acquired, err := r.rdb[0].SetNX(ctx, key, "locked", time.Duration(ttlSec)*time.Second).Result()
	if err != nil {
		return false, err
	}
	return acquired, nil
}

func (r *RedisCache) unlockMutex(ctx context.Context, key string) error {
	_, err := r.rdb[0].Del(ctx, key).Result()
	return err
}

func (r *RedisCache) Mutex(ctx context.Context, key string, ttlSec int, timeoutSec int64, cb func()) error {
	startTs := time.Now().Unix()
	for {
		acquired, err := r.lockMutex(ctx, key, ttlSec)
		if err != nil {
			// With timeout=0, fail immediately on any error
			if timeoutSec == 0 {
				return fmt.Errorf("error while locking mutex %s: %w", key, err)
			}
			// With timeout>0, warn but continue retrying (transient errors)
			common.L.Warn(
				fmt.Sprintf("error while locking mutex %s: %s", key, err.Error()), common.F(ctx)...)
		}
		if acquired {
			common.L.Debug(fmt.Sprintf("mutex locked: %s", key), common.F(ctx)...)
			break
		}

		// For timeout=0, fail immediately if lock is already held
		if timeoutSec == 0 {
			return fmt.Errorf("lock already held by another coordinator for mutex %s", key)
		}

		// Check if we've exceeded the timeout (only for positive timeouts)
		nowTs := time.Now().Unix()
		diffSec := nowTs - startTs
		if timeoutSec > 0 && diffSec >= timeoutSec {
			return fmt.Errorf("timeout while locking mutex %s after %d seconds", key, diffSec)
		}

		time.Sleep(1 * time.Second)
	}

	// Ensure unlock happens even if callback panics
	defer func() {
		err := r.unlockMutex(ctx, key)
		if err != nil {
			common.L.Warn(
				fmt.Sprintf("error while unlocking mutex %s: %s", key, err.Error()), common.F(ctx)...)
		} else {
			common.L.Debug(fmt.Sprintf("mutex unlocked: %s", key), common.F(ctx)...)
		}
	}()

	cb()
	return nil
}

// MultiMutex acquires multiple locks in a deterministic order (sorted keys) to prevent deadlocks.
// All locks must be acquired before the callback is executed.
// If any lock acquisition times out, all previously acquired locks are released.
func (r *RedisCache) MultiMutex(ctx context.Context, keys []string, ttlSec int, timeoutSec int64, cb func()) error {
	if len(keys) == 0 {
		cb()
		return nil
	}

	// Sort keys to ensure consistent lock ordering and prevent deadlocks
	sortedKeys := make([]string, len(keys))
	copy(sortedKeys, keys)
	sortStrings(sortedKeys)

	// Remove duplicates
	uniqueKeys := make([]string, 0, len(sortedKeys))
	for i, k := range sortedKeys {
		if i == 0 || k != sortedKeys[i-1] {
			uniqueKeys = append(uniqueKeys, k)
		}
	}

	common.L.Debug(fmt.Sprintf("MultiMutex: acquiring %d locks: %v", len(uniqueKeys), uniqueKeys), common.F(ctx)...)

	// Track acquired locks for cleanup on failure
	acquiredLocks := make([]string, 0, len(uniqueKeys))

	// Helper to release all acquired locks
	releaseAll := func() {
		for i := len(acquiredLocks) - 1; i >= 0; i-- {
			key := acquiredLocks[i]
			if err := r.unlockMutex(ctx, key); err != nil {
				common.L.Warn(fmt.Sprintf("error while unlocking mutex %s during cleanup: %s", key, err.Error()), common.F(ctx)...)
			} else {
				common.L.Debug(fmt.Sprintf("mutex unlocked (cleanup): %s", key), common.F(ctx)...)
			}
		}
	}

	startTs := time.Now().Unix()

	// Acquire locks in order
	for _, key := range uniqueKeys {
		for {
			acquired, err := r.lockMutex(ctx, key, ttlSec)
			if err != nil {
				if timeoutSec == 0 {
					releaseAll()
					return fmt.Errorf("error while locking mutex %s: %w", key, err)
				}
				common.L.Warn(fmt.Sprintf("error while locking mutex %s: %s", key, err.Error()), common.F(ctx)...)
			}

			if acquired {
				common.L.Debug(fmt.Sprintf("mutex locked: %s", key), common.F(ctx)...)
				acquiredLocks = append(acquiredLocks, key)
				break
			}

			if timeoutSec == 0 {
				releaseAll()
				return fmt.Errorf("lock already held by another coordinator for mutex %s", key)
			}

			nowTs := time.Now().Unix()
			diffSec := nowTs - startTs
			if timeoutSec > 0 && diffSec >= timeoutSec {
				releaseAll()
				return fmt.Errorf("timeout while locking mutex %s after %d seconds", key, diffSec)
			}

			time.Sleep(100 * time.Millisecond) // Faster polling for multi-lock scenarios
		}
	}

	common.L.Debug(fmt.Sprintf("MultiMutex: all %d locks acquired", len(uniqueKeys)), common.F(ctx)...)

	// Ensure all locks are released even if callback panics
	defer func() {
		releaseAll()
		common.L.Debug(fmt.Sprintf("MultiMutex: all %d locks released", len(uniqueKeys)), common.F(ctx)...)
	}()

	cb()
	return nil
}

// sortStrings sorts a slice of strings in place (simple bubble sort for small slices)
func sortStrings(s []string) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// AcquireMultiLock acquires multiple distributed locks in a deterministic order (sorted keys).
// Unlike MultiMutex, this does NOT automatically release locks - caller must call ReleaseMultiLock.
// The TTL acts as a safety net: if the caller crashes or forgets to release, locks auto-expire.
func (r *RedisCache) AcquireMultiLock(ctx context.Context, keys []string, ttlSec int, timeoutSec int64) error {
	if len(keys) == 0 {
		return nil
	}

	// Sort keys to ensure consistent lock ordering and prevent deadlocks
	sortedKeys := make([]string, len(keys))
	copy(sortedKeys, keys)
	sortStrings(sortedKeys)

	// Remove duplicates
	uniqueKeys := make([]string, 0, len(sortedKeys))
	for i, k := range sortedKeys {
		if i == 0 || k != sortedKeys[i-1] {
			uniqueKeys = append(uniqueKeys, k)
		}
	}

	common.L.Debug(fmt.Sprintf("AcquireMultiLock: acquiring %d locks: %v", len(uniqueKeys), uniqueKeys), common.F(ctx)...)

	// Track acquired locks for cleanup on failure
	acquiredLocks := make([]string, 0, len(uniqueKeys))

	// Helper to release all acquired locks on failure
	releaseAll := func() {
		for i := len(acquiredLocks) - 1; i >= 0; i-- {
			key := acquiredLocks[i]
			if err := r.unlockMutex(ctx, key); err != nil {
				common.L.Warn(fmt.Sprintf("AcquireMultiLock: error releasing lock %s during cleanup: %s", key, err.Error()), common.F(ctx)...)
			} else {
				common.L.Debug(fmt.Sprintf("AcquireMultiLock: lock released (cleanup): %s", key), common.F(ctx)...)
			}
		}
	}

	startTs := time.Now().Unix()

	// Acquire locks in order
	for _, key := range uniqueKeys {
		for {
			acquired, err := r.lockMutex(ctx, key, ttlSec)
			if err != nil {
				if timeoutSec == 0 {
					releaseAll()
					return fmt.Errorf("AcquireMultiLock: error while locking %s: %w", key, err)
				}
				common.L.Warn(fmt.Sprintf("AcquireMultiLock: error while locking %s: %s", key, err.Error()), common.F(ctx)...)
			}

			if acquired {
				common.L.Debug(fmt.Sprintf("AcquireMultiLock: lock acquired: %s", key), common.F(ctx)...)
				acquiredLocks = append(acquiredLocks, key)
				break
			}

			if timeoutSec == 0 {
				releaseAll()
				return fmt.Errorf("AcquireMultiLock: lock already held for %s", key)
			}

			nowTs := time.Now().Unix()
			diffSec := nowTs - startTs
			if timeoutSec > 0 && diffSec >= timeoutSec {
				releaseAll()
				return fmt.Errorf("AcquireMultiLock: timeout while locking %s after %d seconds", key, diffSec)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	common.L.Debug(fmt.Sprintf("AcquireMultiLock: all %d locks acquired successfully", len(uniqueKeys)), common.F(ctx)...)
	return nil
}

// ReleaseMultiLock releases multiple previously acquired locks.
// Safe to call even if some locks have already expired or were never acquired.
func (r *RedisCache) ReleaseMultiLock(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	common.L.Debug(fmt.Sprintf("ReleaseMultiLock: releasing %d locks: %v", len(keys), keys), common.F(ctx)...)

	var lastErr error
	for _, key := range keys {
		if err := r.unlockMutex(ctx, key); err != nil {
			common.L.Warn(fmt.Sprintf("ReleaseMultiLock: error releasing lock %s: %s", key, err.Error()), common.F(ctx)...)
			lastErr = err
		} else {
			common.L.Debug(fmt.Sprintf("ReleaseMultiLock: lock released: %s", key), common.F(ctx)...)
		}
	}

	return lastErr
}

func (r *RedisCache) cacheRead(
	ctx context.Context,
	db int,
	key string,
	ttlSec uint,
	missCb func() (string, error),
	hitCb func(string) (bool, string, bool, error),
) (string, error) {
	value, err := r.Get(ctx, db, key)
	if err != nil {
		return "", err
	}
	if len(value) != 0 {
		if hitCb != nil {
			readOnly, newValue, invalidate, err := hitCb(value)
			if err != nil {
				return "", err
			}
			if !readOnly {
				if invalidate {
					r.Del(ctx, db, key)
				}
				if value != newValue {
					err := r.Set(ctx, db, key, value, ttlSec)
					if err != nil {
						return "", err
					}
				}
			}
		}
		return value, nil
	}
	value, err = missCb()
	if err != nil {
		return "", err
	}
	err = r.Set(ctx, db, key, value, ttlSec)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *RedisCache) CacheRead(
	ctx context.Context,
	db int,
	key string,
	ttlSec uint,
	missCb func() (string, error),
) (string, error) {
	return r.cacheRead(ctx, db, key, ttlSec, missCb, nil)
}

func (r *RedisCache) CacheRead2(
	ctx context.Context,
	key string,
	missCb func() (string, error),
) (string, error) {
	return r.CacheRead(ctx, 0, key, 0, missCb)
}

func (r *RedisCache) CacheRead3(
	ctx context.Context,
	db int,
	key string,
	ttlSec uint,
	missCb func() (string, error),
	hitCb func(string) (bool, string, bool, error),
) (string, error) {
	return r.cacheRead(ctx, db, key, ttlSec, missCb, hitCb)
}

func (r *RedisCache) Keys(ctx context.Context, db int, pattern string) ([]string, error) {
	val, err := r.rdb[db].Keys(ctx, pattern).Result()
	if err != nil {
		common.L.Warn(err.Error(), common.F(ctx)...)
		return nil, err
	}
	return val, nil
}
