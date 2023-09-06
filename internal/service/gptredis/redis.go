package gptredis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	_openai "github.com/sashabaranov/go-openai"
	"openai/internal/config"
	"openai/internal/util"
	"strconv"
	"time"
)

var ctx = context.Background()
var rdb, brotherRdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     config.C.Redis.Addr,
		Password: "", // no password set
		DB:       config.C.Redis.DB,
	})
	brotherRdb = redis.NewClient(&redis.Options{
		Addr:     config.C.Redis.Addr,
		Password: "", // no password set
		DB:       config.C.Redis.BrotherDB,
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
	return rdb.Set(ctx, buildReplyKey(msgId), reply, time.Hour*24*7).Err()
}

func AppendReplyChunk(msgId int64, chunk string) error {
	err := rdb.RPush(ctx, buildReplyChunksKey(msgId), chunk).Err()
	if err != nil {
		return err
	}
	err = rdb.Expire(ctx, buildReplyChunksKey(msgId), time.Hour*24).Err()
	return err
}

func ReplyChunksExists(msgId int64) (bool, error) {
	code, err := rdb.Exists(ctx, buildReplyChunksKey(msgId)).Result()
	return code == 1, err
}

func GetLengthOfReplyChunks(msgId int64) (int64, error) {
	return rdb.LLen(ctx, buildReplyChunksKey(msgId)).Result()
}

func GetReplyChunks(msgId int64, from int64, to int64) ([]string, error) {
	return rdb.LRange(ctx, buildReplyChunksKey(msgId), from, to).Result()
}

func buildReplyChunksKey(msgId int64) string {
	return "msg-id:" + strconv.FormatInt(msgId, 10) + ":reply-chunks"
}

func SetMessages(toUserName string, messages []_openai.ChatCompletionMessage) error {
	newRoundsStr, err := util.StringifyMessages(messages)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, buildMessagesKey(toUserName), newRoundsStr, time.Minute*5).Err()
}

func FetchMessages(toUserName string) ([]_openai.ChatCompletionMessage, error) {
	var messages []_openai.ChatCompletionMessage
	messagesStr, err := rdb.Get(ctx, buildMessagesKey(toUserName)).Result()
	if err != nil {
		if err == redis.Nil {
			return messages, nil
		}
		return nil, err
	}
	messages, err = util.ParseMessages(messagesStr)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func DelMessages(toUserName string) error {
	return rdb.Del(ctx, buildMessagesKey(toUserName)).Err()
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

func FetchMediaId(imageName string) (string, error) {
	return rdb.Get(ctx, getMediaIdKey(imageName)).Result()
}

func SetMediaId(mediaId string, mediaName string, expiration time.Duration) error {
	return rdb.Set(ctx, getMediaIdKey(mediaName), mediaId, expiration).Err()
}

func getMediaIdKey(mediaName string) string {
	return fmt.Sprintf("media-id-of-%s", mediaName)
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

func buildCodeKey(code string) string {
	return "code:" + code
}

func SetCodeDetail(code string, codeDetail string, useBrotherDB bool) error {
	myRdb := rdb
	if useBrotherDB {
		myRdb = brotherRdb
	}
	return myRdb.Set(ctx, buildCodeKey(code), codeDetail, 0).Err()
}

func FetchCodeDetail(code string) (string, error) {
	return rdb.Get(ctx, buildCodeKey(code)).Result()
}

func SetPaidBalance(user string, balance int) error {
	return rdb.Set(ctx, buildPaidBalance(user), balance, 0).Err()
}

func FetchPaidBalance(user string) (int, error) {
	balanceStr, err := rdb.Get(ctx, buildPaidBalance(user)).Result()
	if err != nil {
		return 0, err
	}
	balance, err := strconv.Atoi(balanceStr)
	return balance, err
}

func DecrPaidBalance(usr string) (int64, error) {
	return rdb.Decr(ctx, buildPaidBalance(usr)).Result()
}

func buildPaidBalance(user string) string {
	return "user:" + user + ":paid-balance"
}

func buildOpenIdKey(authCode string) string {
	return "auth-code:" + authCode + ":open-id"
}

func FetchOpenId(authCode string) (string, error) {
	return rdb.Get(ctx, buildOpenIdKey(authCode)).Result()
}

func SetOpenId(authCode string, openId string) error {
	return rdb.Set(ctx, buildOpenIdKey(authCode), openId, time.Hour*12).Err()
}

func buildQuotaKey(user string, day string) string {
	return "user:" + user + ":day:" + day + ":quota"
}

func SetQuota(user string, day string, quota int) error {
	return rdb.Set(ctx, buildQuotaKey(user, day), quota, time.Hour*24).Err()
}

func GetQuota(user string, day string) (int, error) {
	quotaStr, err := rdb.Get(ctx, buildQuotaKey(user, day)).Result()
	if err != nil {
		return 0, err
	}
	quota, err := strconv.Atoi(quotaStr)
	return quota, err
}
