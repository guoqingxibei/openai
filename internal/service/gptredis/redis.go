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

func FetchReply(msgId int64) (string, error) {
	reply, err := rdb.Get(ctx, buildReplyKey(msgId)).Result()
	if err == nil {
		return reply, nil
	}
	return "", err
}

func SetReply(msgId int64, reply string) error {
	return rdb.Set(ctx, buildReplyKey(msgId), reply, time.Hour*24).Err()
}

func SetMessages(toUserName string, messages []openai.Message) error {
	newRoundsStr, err := openai.StringifyMessages(messages)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, buildMessagesKey(toUserName), newRoundsStr, time.Minute*5).Err()
}

func FetchMessages(toUserName string) ([]openai.Message, error) {
	var messages []openai.Message
	messagesStr, err := rdb.Get(ctx, buildMessagesKey(toUserName)).Result()
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
	return rdb.Del(ctx, buildReplyKey(msgId)).Err()
}

func buildReplyKey(msgId int64) string {
	return "msg-id:" + strconv.FormatInt(msgId, 10) + ":reply"
}

func IncAccessTimes(msgId int64) (int64, error) {
	msgIdStr := strconv.FormatInt(msgId, 10)
	key := buildAccessTimes(msgIdStr)
	times, err := rdb.Incr(ctx, key).Result()
	if times == 1 {
		rdb.Expire(ctx, key, time.Second*30)
	}
	if err != nil {
		return 0, nil
	}
	return times, nil
}

func buildAccessTimes(msgIdStr string) string {
	return "msg-id:" + msgIdStr + ":access-times"
}

func FetchBaiduApiAccessToken() (string, error) {
	return rdb.Get(ctx, getBaiduApiAccessTokenKey()).Result()
}

func SetBaiduApiAccessToken(accessToken string, expiration time.Duration) error {
	return rdb.Set(ctx, getBaiduApiAccessTokenKey(), accessToken, expiration).Err()
}

func getBaiduApiAccessTokenKey() string {
	return "baidu-api-access-token"
}

func FetchWechatApiAccessToken() (string, error) {
	return rdb.Get(ctx, getWechatApiAccessTokenKey()).Result()
}

func SetWechatApiAccessToken(accessToken string, expiration time.Duration) error {
	return rdb.Set(ctx, getWechatApiAccessTokenKey(), accessToken, expiration).Err()
}

func getWechatApiAccessTokenKey() string {
	return "wechat-api-access-token"
}

func SetModeForUser(user string, mode string) error {
	return rdb.Set(ctx, buildModeKey(user), mode, 0).Err()
}

func FetchModeForUser(user string) (string, error) {
	return rdb.Get(ctx, buildModeKey(user)).Result()
}

func buildModeKey(user string) string {
	return "user:" + user + ":mode"
}

func FetchBalance(user string, mode string, day string) (int, error) {
	balance, err := rdb.Get(ctx, buildBalanceKey(user, mode, day)).Result()
	cnt, _ := strconv.Atoi(balance)
	return cnt, err
}

func SetBalance(user string, mode string, day string, balance int) error {
	return rdb.Set(ctx, buildBalanceKey(user, mode, day), strconv.Itoa(balance), time.Hour*24).Err()
}

func DecrBalance(user string, mode string, day string) (int, error) {
	balance, err := rdb.Decr(ctx, buildBalanceKey(user, mode, day)).Result()
	return int(balance), err
}

func buildBalanceKey(user string, mode string, day string) string {
	return "user:" + user + ":mode:" + mode + ":day:" + day + ":balance"
}

func FetchMediaIdOfDonateQr() (string, error) {
	return rdb.Get(ctx, getMediaIdKey()).Result()
}

func SetMediaIdOfDonateQr(mediaId string, expiration time.Duration) error {
	return rdb.Set(ctx, getMediaIdKey(), mediaId, expiration).Err()
}

func getMediaIdKey() string {
	return "media-id-of-donate-qr"
}

func buildUsageKey(user string) string {
	return "user:" + user + ":used-times"
}

func IncUsedTimes(user string) (int, error) {
	times, err := rdb.Incr(ctx, buildUsageKey(user)).Result()
	return int(times), err
}

func buildSubscribeTimestampKey(user string) string {
	return "user:" + user + ":subscribe-timestamp"
}

func SetSubscribeTimestamp(user string, timestamp int64) error {
	return rdb.Set(ctx, buildSubscribeTimestampKey(user), strconv.FormatInt(timestamp, 10), 0).Err()
}

func FetchSubscribeTimestamp(user string) (int64, error) {
	timestampStr, err := rdb.Get(ctx, buildSubscribeTimestampKey(user)).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(timestampStr, 10, 64)
}
