package cache

import (
	"context"
)

func Init(ctx context.Context) {
	redisCache := NewRedisCache()
	err := redisCache.Init(ctx)
	if err != nil {
		panic(err)
	}
	SetDefaultCache(redisCache)
}

func Exists(ctx context.Context, db int, key string) (bool, error) {
	return defaultCache.Exists(ctx, db, key)
}

func Get(ctx context.Context, db int, key string) (string, error) {
	return defaultCache.Get(ctx, db, key)
}

func Set(ctx context.Context, db int, key, value string, ttlSec uint) error {
	return defaultCache.Set(ctx, db, key, value, ttlSec)
}

func Del(ctx context.Context, db int, key string) error {
	return defaultCache.Del(ctx, db, key)
}

func ListPushLeft(ctx context.Context, db int, key, element string) error {
	return defaultCache.ListPushLeft(ctx, db, key, element)
}

func ListPopRight(ctx context.Context, db int, key string) (string, error) {
	return defaultCache.ListPopRight(ctx, db, key)
}

func ListLen(ctx context.Context, db int, key string) (int64, error) {
	return defaultCache.ListLen(ctx, db, key)
}

func ListRange(ctx context.Context, db int, key string, start, stop int64) ([]string, error) {
	return defaultCache.ListRange(ctx, db, key, start, stop)
}

func AddToSet(ctx context.Context, db int, key, element string) error {
	return defaultCache.AddToSet(ctx, db, key, element)
}

func IsSetMember(ctx context.Context, db int, key, element string) (bool, error) {
	return defaultCache.IsSetMember(ctx, db, key, element)
}

func SetMembers(ctx context.Context, db int, key string) ([]string, error) {
	return defaultCache.SetMembers(ctx, db, key)
}

func RPush(ctx context.Context, db int, key, element string) error {
	return defaultCache.RPush(ctx, db, key, element)
}

func HSet(ctx context.Context, db int, key, field, value string) error {
	return defaultCache.HSet(ctx, db, key, field, value)
}

func HGet(ctx context.Context, db int, key, field string) (string, error) {
	return defaultCache.HGet(ctx, db, key, field)
}

func HGetAll(ctx context.Context, db int, key string) (map[string]string, error) {
	return defaultCache.HGetAll(ctx, db, key)
}

func HDel(ctx context.Context, db int, key string, fields ...string) error {
	return defaultCache.HDel(ctx, db, key, fields...)
}

func HMSet(ctx context.Context, db int, key string, fields map[string]string) error {
	return defaultCache.HMSet(ctx, db, key, fields)
}

func ZAdd(ctx context.Context, db int, key string, score float64, member string) error {
	return defaultCache.ZAdd(ctx, db, key, score, member)
}

func ZRem(ctx context.Context, db int, key string, members ...string) error {
	return defaultCache.ZRem(ctx, db, key, members...)
}

func ZRangeByScore(ctx context.Context, db int, key string, min, max string, offset, count int64) ([]string, error) {
	return defaultCache.ZRangeByScore(ctx, db, key, min, max, offset, count)
}

func ZRevRangeByScore(ctx context.Context, db int, key string, max, min string, offset, count int64) ([]string, error) {
	return defaultCache.ZRevRangeByScore(ctx, db, key, max, min, offset, count)
}

func ZCard(ctx context.Context, db int, key string) (int64, error) {
	return defaultCache.ZCard(ctx, db, key)
}

func ZRangeWithScores(ctx context.Context, db int, key string, start, stop int64) ([]ZMember, error) {
	return defaultCache.ZRangeWithScores(ctx, db, key, start, stop)
}

func ZRevRangeWithScores(ctx context.Context, db int, key string, start, stop int64) ([]ZMember, error) {
	return defaultCache.ZRevRangeWithScores(ctx, db, key, start, stop)
}

func ZRemRangeByScore(ctx context.Context, db int, key, min, max string) error {
	return defaultCache.ZRemRangeByScore(ctx, db, key, min, max)
}

func JsonSet(ctx context.Context, db int, key, path, value string) error {
	return defaultCache.JsonSet(ctx, db, key, path, value)
}

func JsonGet(ctx context.Context, db int, key, path string) (string, error) {
	return defaultCache.JsonGet(ctx, db, key, path)
}

func JsonArrAppend(ctx context.Context, db int, key, path, element string) error {
	return defaultCache.JsonArrAppend(ctx, db, key, path, element)
}

func Keys(ctx context.Context, db int, pattern string) ([]string, error) {
	return defaultCache.Keys(ctx, db, pattern)
}
