package redis

import (
	"be-cleverschool/config"
	"be-cleverschool/database/db"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedResult[T any] struct {
	Entities   []T   `json:"entities"`
	TotalCount int64 `json:"total_count"`
}

func RememberCache[T any](cacheKey string, ttl time.Duration, fetchFunc func() (T, error)) (T, error) {
	var cachedData T

	val, err := db.RedisClient.Get(db.Ctx, cacheKey).Result()
	if err == nil {
		err := json.Unmarshal([]byte(val), &cachedData)
		if err == nil {
			return cachedData, nil
		}
	} else if err != redis.Nil {
		return *new(T), fmt.Errorf("failed to get cache: %w", err)
	}

	data, err := fetchFunc()
	if err != nil {
		return *new(T), fmt.Errorf("failed to fetch data: %w", err)
	}

	jsonData, err := json.Marshal(data)
	if err == nil {
		err = db.RedisClient.Set(db.Ctx, cacheKey, jsonData, ttl).Err()

		if err != nil {
			return *new(T), fmt.Errorf("failed to set cache: %w", err)
		}
	}

	return data, nil
}

func DeleteCache(cacheKey string) error {
	err := db.RedisClient.Del(db.Ctx, cacheKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache key '%s': %w", cacheKey, err)
	}
	return nil
}

func GetFromCacheByKey[T any](result *CachedResult[T], params map[string]interface{}) error {
	cacheKey := GenerateKeyFromParams[T](params)

	val, err := db.RedisClient.Get(db.Ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	err = json.Unmarshal([]byte(val), result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return nil
}

func SetToCacheByKey[T any](result CachedResult[T], params map[string]interface{}, ttl time.Duration) error {
	params["_group"] = "ApiList"

	_ = CheckRedisMemoryAndClean("cache:ApiList", 1024)

	cacheKey := GenerateKeyFromParams[T](params)

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal data for cache: %w", err)
	}

	err = db.RedisClient.Set(db.Ctx, cacheKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

func DeleteCacheByKey[T any]() error {
	typeName := fmt.Sprintf("%T", new(T))
	typeName = strings.Replace(typeName, "*", "", -1)

	err := ClearCacheByPrefix(typeName)

	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}
	return nil
}

func SetToCacheById[T any](id int, entities T, ttl time.Duration) error {
	cacheKey := GetKeyById[T](id)

	data, err := json.Marshal(entities)
	if err != nil {
		return fmt.Errorf("failed to marshal entities: %w", err)
	}

	err = db.RedisClient.Set(db.Ctx, cacheKey, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache for key '%s': %w", cacheKey, err)
	}
	return nil
}

func GetFromCacheById[T any](id int) (T, error) {
	cacheKey := GetKeyById[T](id)

	result, err := db.RedisClient.Get(db.Ctx, cacheKey).Result()

	if err == redis.Nil {
		var emptyEntity T
		return emptyEntity, nil
	} else if err != nil {
		return *new(T), fmt.Errorf("failed to get cache for key '%s': %w", cacheKey, err)
	}

	var entity T
	err = json.Unmarshal([]byte(result), &entity)
	if err != nil {
		return *new(T), fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return entity, nil
}

func DeleteCacheById[T any](id int) error {
	cacheKey := GetKeyById[T](id)

	err := db.RedisClient.Del(db.Ctx, cacheKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache key '%s': %w", cacheKey, err)
	}
	return nil
}

func GetKeyByFilter[T any](filter map[string]interface{}, page int, limit int) string {
	sortedKeys := make([]string, 0, len(filter))
	for key := range filter {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)

	sortedFilter := make(map[string]interface{})
	for _, key := range sortedKeys {
		sortedFilter[key] = filter[key]
	}

	filterBytes, err := json.Marshal(sortedFilter)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	typeName := fmt.Sprintf("%T", new(T))
	typeName = strings.Replace(typeName, "*", "", -1)

	cacheKey := fmt.Sprintf("%s:%s:%d:%d", typeName, string(filterBytes), page, limit)
	return cacheKey
}

func GetKeyById[T any](id int) string {
	typeName := fmt.Sprintf("%T", new(T))
	typeName = strings.Replace(typeName, "*", "", -1)
	cacheKey := fmt.Sprintf("%s:id:%d", typeName, id)

	return cacheKey
}

func ClearCacheByPrefix(prefix string) error {
	ctx := context.Background()
	var cursor uint64
	var deletedCount int64
	var consecutiveEmptyScans int

	// Check if Redis client is available
	if db.RedisClient == nil {
		return fmt.Errorf("Redis client is nil - Redis may not be connected")
	}

	for {
		keys, cursor, err := db.RedisClient.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			config.Log.Errorf("Redis scan error for prefix '%s': %v", prefix, err)
			return fmt.Errorf("failed to scan keys with prefix '%s': %w", prefix, err)
		}

		if len(keys) > 0 {
			deleted, err := db.RedisClient.Del(ctx, keys...).Result()
			if err != nil {
				config.Log.Errorf("Redis delete error for prefix '%s': %v", prefix, err)
				return fmt.Errorf("failed to delete keys with prefix '%s': %w", prefix, err)
			}
			deletedCount += deleted
			consecutiveEmptyScans = 0
		} else {
			consecutiveEmptyScans++
		}

		if cursor == 0 {
			break
		}

		if consecutiveEmptyScans >= 1 {
			break
		}
	}

	return nil
}

func ClearAllCache() error {
	err := db.RedisClient.FlushDB(db.Ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to clear Redis cache: %w", err)
	}
	return nil
}

func GenerateKeyFromParams[T any](params map[string]interface{}) string {
	bytes, _ := json.Marshal(params)
	hash := sha1.Sum(bytes)

	typeName := fmt.Sprintf("%T", new(T))
	typeName = strings.Replace(typeName, "*", "", -1)

	return fmt.Sprintf("%s:%x", typeName, hash)
}

func CheckRedisMemoryAndClean(groupPrefix string, maxMemoryUsedMB int) error {
	info, err := db.RedisClient.Info(db.Ctx, "memory").Result()
	if err != nil {
		return fmt.Errorf("failed to get Redis memory info: %w", err)
	}

	var usedMemoryBytes int64
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "used_memory:") {
			fmt.Sscanf(line, "used_memory:%d", &usedMemoryBytes)
			break
		}
	}

	usedMB := usedMemoryBytes / (1024 * 1024)
	if usedMB >= int64(maxMemoryUsedMB) {
		config.Log.Warnf("Redis memory usage is %d MB, exceeding %d MB — cleaning keys with prefix '%s'", usedMB, maxMemoryUsedMB, groupPrefix)
		return ClearCacheByPrefix(groupPrefix)
	}

	return nil
}

