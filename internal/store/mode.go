package store

import (
	"errors"
	"github.com/redis/go-redis/v9"
	"openai/internal/constant"
)

func buildModeKey(user string) string {
	return "user:" + user + ":mode"
}

func GetMode(user string) (string, error) {
	result, err := client.Get(ctx, buildModeKey(user)).Result()
	if errors.Is(err, redis.Nil) {
		return constant.GPT3, nil
	}
	return result, err
}

func SetMode(user string, mode string) error {
	return client.Set(ctx, buildModeKey(user), mode, 0).Err()
}
