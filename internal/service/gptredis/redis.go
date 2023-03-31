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

func set(key string, value string, expiration time.Duration) error {
	return rdb.Set(ctx, key, value, expiration).Err()
}

func get(key string) (string, error) {
	return rdb.Get(ctx, key).Result()
}

func del(key string) error {
	return rdb.Del(ctx, key).Err()
}

func inc(key string) (int64, error) {
	return rdb.Incr(ctx, key).Result()
}

func FetchReply(msgId int64) (string, error) {
	reply, err := get(buildReplyKey(msgId))
	if err == nil {
		return reply, nil
	}
	return "", err
}

func SetReply(msgId int64, reply string) error {
	return set(buildReplyKey(msgId), reply, time.Hour*24)
}

func SetMessages(toUserName string, messages []openai.Message) error {
	newRoundsStr, err := openai.StringifyMessages(messages)
	if err != nil {
		return err
	}
	return set(buildMessagesKey(toUserName), newRoundsStr, time.Minute*5)
}

func FetchMessages(toUserName string) ([]openai.Message, error) {
	var messages []openai.Message
	messagesStr, err := get(buildMessagesKey(toUserName))
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

func DelReply(msgId int64) error {
	return del(buildReplyKey(msgId))
}

func buildReplyKey(msgId int64) string {
	return "msg-id:" + strconv.FormatInt(msgId, 10) + ":reply"
}

func IncAccessTimes(msgId int64) (int64, error) {
	msgIdStr := strconv.FormatInt(msgId, 10)
	times, err := inc("msg-id:" + msgIdStr + ":access-times")
	if err != nil {
		return 0, nil
	}
	return times, nil
}

func FetchBaiduApiAccessToken() (string, error) {
	return get(getBaiduApiAccessTokenKey())
}

func SetBaiduApiAccessToken(accessToken string, expiration time.Duration) error {
	return set(getBaiduApiAccessTokenKey(), accessToken, expiration)
}

func getBaiduApiAccessTokenKey() string {
	return "baidu-api-access-token"
}

func FetchWechatApiAccessToken() (string, error) {
	return get(getWechatApiAccessTokenKey())
}

func SetWechatApiAccessToken(accessToken string, expiration time.Duration) error {
	return set(getWechatApiAccessTokenKey(), accessToken, expiration)
}

func getWechatApiAccessTokenKey() string {
	return "wechat-api-access-token"
}

func SetModeForUser(user string, mode string) error {
	return set(buildModeKey(user), mode, 0)
}

func FetchModeForUser(user string) (string, error) {
	return get(buildModeKey(user))
}

func buildModeKey(user string) string {
	return "user:" + user + ":mode"
}

func FetchImageBalance(user string) (int, error) {
	balance, err := get(buildImageBalanceKey(user))
	cnt, _ := strconv.Atoi(balance)
	return cnt, err
}

func SetImageBalance(user string, balance int) error {
	return set(buildImageBalanceKey(user), strconv.Itoa(balance), time.Hour*24)
}

func DecrImageBalance(user string) (int, error) {
	balance, err := rdb.Decr(ctx, buildImageBalanceKey(user)).Result()
	return int(balance), err
}

func buildImageBalanceKey(user string) string {
	return "user:" + user + ":image-balance"
}

func FetchMediaIdOfDonateQr() (string, error) {
	return get(getMediaIdKey())
}

func SetMediaIdOfDonateQr(mediaId string, expiration time.Duration) error {
	return set(getMediaIdKey(), mediaId, expiration)
}

func getMediaIdKey() string {
	return "media-id-of-donate-qr"
}

func buildUsageKey(user string) string {
	return "user:" + user + ":used-times"
}

func IncUsedTimes(user string) (int, error) {
	times, err := inc(buildUsageKey(user))
	return int(times), err
}
