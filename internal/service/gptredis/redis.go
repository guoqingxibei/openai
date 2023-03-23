package gptredis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"openai/internal/config"
	"openai/internal/service/openai"
	"strconv"
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

func FetchReply(shortMsgId string) (string, error) {
	reply, err := Get(buildReplyKey(shortMsgId))
	if err == nil {
		return reply, nil
	}
	return "", err
}

func SetReply(shortMsgId string, reply string) error {
	return Set(buildReplyKey(shortMsgId), reply, time.Hour*24*7)
}

func SetMessages(toUserName string, messages []openai.Message) error {
	newRoundsStr, err := openai.StringifyMessages(messages)
	if err != nil {
		return err
	}
	return Set(buildMessagesKey(toUserName), newRoundsStr, time.Minute*10)
}

func FetchMessages(toUserName string) ([]openai.Message, error) {
	var messages []openai.Message
	messagesStr, err := Get(buildMessagesKey(toUserName))
	if err != nil {
		if err == redis.Nil {
			return messages, nil
		}
		return nil, err
	}
	messages, err = openai.ParseMessages(messagesStr)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func buildMessagesKey(toUserName string) string {
	return "user:" + toUserName + ":messages"
}

func DelReply(shortMsgId string) error {
	return Del(buildReplyKey(shortMsgId))
}

func buildReplyKey(shortMsgId string) string {
	return "short-msg-id:" + shortMsgId + ":reply"
}

func generateShortMsgId() (string, error) {
	shortMsgId, err := Inc("current-max-short-id")
	if err == nil {
		return strconv.FormatInt(shortMsgId, 10), nil
	}
	return "", err
}

func FetchShortMsgId(longMsgId string) (string, error) {
	key := buildShortMsgIdKey(longMsgId)
	shortMsgId, err := Get(key)
	if err == nil {
		return shortMsgId, nil
	}
	if err == redis.Nil {
		shortMsgId, err := generateShortMsgId()
		if err == nil {
			err := Set(key, shortMsgId, time.Hour*24*7)
			if err == nil {
				return shortMsgId, nil
			}
			return "", err
		}
		return "", err
	}
	return "", err
}

func buildShortMsgIdKey(longMsgId string) string {
	return "long-msg-id:" + longMsgId + ":short-msg-id"
}

func IncAccessTimes(shortMsgId string) (int64, error) {
	times, err := Inc("short-msg-id:" + shortMsgId + ":access-times")
	if err != nil {
		return 0, nil
	}
	return times, nil
}

func FetchBaiduApiAccessToken() (string, error) {
	return Get(getBaiduApiAccessTokenKey())
}

func SetBaiduApiAccessToken(accessToken string, expiration time.Duration) error {
	return Set(getBaiduApiAccessTokenKey(), accessToken, expiration)
}

func getBaiduApiAccessTokenKey() string {
	return "baidu-api-access-token"
}
