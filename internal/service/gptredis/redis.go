package gptredis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"openai/internal/config"
	"time"
)

var ctx = context.Background()
var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     config.C.Redis.Addr,
		Password: "", // no password set
		DB:       config.C.Redis.DB,
	})
}

func Set(key string, value string, expiration time.Duration) error {
	return rdb.Set(ctx, key, value, expiration).Err()
}

func Get(key string) (string, error) {
	return rdb.Get(ctx, key).Result()
}

func Del(key string) error {
	return rdb.Del(ctx, key).Err()
}

func Inc(key string) (int64, error) {
	return rdb.Incr(ctx, key).Result()
}
