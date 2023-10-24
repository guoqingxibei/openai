package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	_openai "github.com/sashabaranov/go-openai"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/model"
	"openai/internal/util"
	"strconv"
	"time"
)

const (
	DAY  = time.Hour * 24
	WEEK = DAY * 7
)

var ctx = context.Background()
var client, uncleClient, brotherClient *redis.Client

func init() {
	brotherClient = redis.NewClient(&redis.Options{
		Addr:     config.C.Redis.Addr,
		Password: "", // no password set
		DB:       config.C.Redis.BrotherDB,
	})
	uncleClient = redis.NewClient(&redis.Options{
		Addr:     config.C.Redis.Addr,
		Password: "", // no password set
		DB:       config.C.Redis.UncleDB,
	})
	client = selectDefaultClient()
}

func selectDefaultClient() *redis.Client {
	if util.AccountIsUncle() {
		return uncleClient
	}
	return brotherClient
}

func AppendReplyChunk(msgId int64, chunk string) error {
	err := client.RPush(ctx, buildReplyChunksKey(msgId), chunk).Err()
	if err != nil {
		return err
	}
	err = client.Expire(ctx, buildReplyChunksKey(msgId), WEEK).Err()
	return err
}

func ReplyChunksExists(msgId int64) (bool, error) {
	code, err := client.Exists(ctx, buildReplyChunksKey(msgId)).Result()
	return code == 1, err
}

func GetReplyChunks(msgId int64, from int64, to int64) ([]string, error) {
	return client.LRange(ctx, buildReplyChunksKey(msgId), from, to).Result()
}

func DelReplyChunks(msgId int64) error {
	return client.Del(ctx, buildReplyChunksKey(msgId)).Err()
}

func buildReplyChunksKey(msgId int64) string {
	return "msg-id:" + strconv.FormatInt(msgId, 10) + ":reply-chunks"
}

func SetMessages(toUserName string, messages []_openai.ChatCompletionMessage) error {
	newRoundsStr, err := util.StringifyMessages(messages)
	if err != nil {
		return err
	}
	return client.Set(ctx, buildMessagesKey(toUserName), newRoundsStr, time.Minute*5).Err()
}

func GetMessages(toUserName string) ([]_openai.ChatCompletionMessage, error) {
	var messages []_openai.ChatCompletionMessage
	messagesStr, err := client.Get(ctx, buildMessagesKey(toUserName)).Result()
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
	return client.Del(ctx, buildMessagesKey(toUserName)).Err()
}

func buildMessagesKey(toUserName string) string {
	return "user:" + toUserName + ":messages"
}

func IncAccessTimes(msgId int64) (int64, error) {
	msgIdStr := strconv.FormatInt(msgId, 10)
	key := buildAccessTimes(msgIdStr)
	times, err := client.Incr(ctx, key).Result()
	if times == 1 {
		client.Expire(ctx, key, time.Second*30)
	}
	if err != nil {
		return 0, nil
	}
	return times, nil
}

func buildAccessTimes(msgIdStr string) string {
	return "msg-id:" + msgIdStr + ":access-times"
}

func GetBaiduApiAccessToken() (string, error) {
	return client.Get(ctx, getBaiduApiAccessTokenKey()).Result()
}

func SetBaiduApiAccessToken(accessToken string, expiration time.Duration) error {
	return client.Set(ctx, getBaiduApiAccessTokenKey(), accessToken, expiration).Err()
}

func getBaiduApiAccessTokenKey() string {
	return "baidu-api-access-token"
}

func GetBalance(user string, day string) (int, error) {
	balance, err := client.Get(ctx, buildBalanceKey(user, day)).Result()
	cnt, _ := strconv.Atoi(balance)
	return cnt, err
}

func SetBalance(user string, day string, balance int) error {
	return client.Set(ctx, buildBalanceKey(user, day), strconv.Itoa(balance), DAY).Err()
}

func DecrBalance(user string, day string) (int, error) {
	balance, err := client.Decr(ctx, buildBalanceKey(user, day)).Result()
	return int(balance), err
}

func buildBalanceKey(user string, day string) string {
	return "user:" + user + ":day:" + day + ":balance"
}

func GetMediaId(imageName string) (string, error) {
	return client.Get(ctx, getMediaIdKey(imageName)).Result()
}

func SetMediaId(mediaId string, mediaName string, expiration time.Duration) error {
	return client.Set(ctx, getMediaIdKey(mediaName), mediaId, expiration).Err()
}

func getMediaIdKey(mediaName string) string {
	return fmt.Sprintf("media-id-of-%s", mediaName)
}

func buildUsageKey(user string) string {
	return "user:" + user + ":used-times"
}

func IncUsedTimes(user string) (int, error) {
	times, err := client.Incr(ctx, buildUsageKey(user)).Result()
	return int(times), err
}

func buildSubscribeTimestampKey(user string) string {
	return "user:" + user + ":subscribe-timestamp"
}

func SetSubscribeTimestamp(user string, timestamp int64) error {
	return client.Set(ctx, buildSubscribeTimestampKey(user), strconv.FormatInt(timestamp, 10), 0).Err()
}

func GetSubscribeTimestamp(user string) (int64, error) {
	timestampStr, err := client.Get(ctx, buildSubscribeTimestampKey(user)).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(timestampStr, 10, 64)
}

func buildCodeKey(code string) string {
	return "code:" + code
}

func SetCodeDetail(code string, codeDetail string, useBrotherDB bool) error {
	myClient := client
	if useBrotherDB {
		myClient = brotherClient
	}
	return myClient.Set(ctx, buildCodeKey(code), codeDetail, 0).Err()
}

func GetCodeDetail(code string) (string, error) {
	return client.Get(ctx, buildCodeKey(code)).Result()
}

func SetPaidBalance(user string, balance int) error {
	return SetPaidBalanceWithDB(user, balance, false)
}

func SetPaidBalanceWithDB(user string, balance int, useUncleDB bool) error {
	myClient := client
	if useUncleDB {
		myClient = uncleClient
	}
	return myClient.Set(ctx, buildPaidBalance(user), balance, 0).Err()
}

func GetPaidBalance(user string) (int, error) {
	return GetPaidBalanceWithDB(user, false)
}

func GetPaidBalanceWithDB(user string, useUncleDB bool) (int, error) {
	myClient := client
	if useUncleDB {
		myClient = uncleClient
	}
	balanceStr, err := myClient.Get(ctx, buildPaidBalance(user)).Result()
	if err != nil {
		return 0, err
	}
	balance, err := strconv.Atoi(balanceStr)
	return balance, err
}

func DecrPaidBalance(user string, decrement int64) (int64, error) {
	return client.DecrBy(ctx, buildPaidBalance(user), decrement).Result()
}

func buildPaidBalance(user string) string {
	return "user:" + user + ":paid-balance"
}

func buildOpenIdKey(authCode string) string {
	return "auth-code:" + authCode + ":open-id"
}

func GetOpenId(authCode string) (string, error) {
	return client.Get(ctx, buildOpenIdKey(authCode)).Result()
}

func SetOpenId(authCode string, openId string) error {
	return client.Set(ctx, buildOpenIdKey(authCode), openId, time.Hour*12).Err()
}

func buildQuotaKey(user string, day string) string {
	return "user:" + user + ":day:" + day + ":quota"
}

func SetQuota(user string, day string, quota int) error {
	return client.Set(ctx, buildQuotaKey(user, day), quota, DAY).Err()
}

func GetQuota(user string, day string) (int, error) {
	quotaStr, err := client.Get(ctx, buildQuotaKey(user, day)).Result()
	if err != nil {
		return 0, err
	}
	quota, err := strconv.Atoi(quotaStr)
	return quota, err
}

func buildTransactionKey(outTradeNo string) (key string) {
	key = "out-trade-no:" + outTradeNo + ":transaction"
	return
}

func SetTransaction(outTradeNo string, transaction model.Transaction) (err error) {
	tranBytes, _ := json.Marshal(transaction)
	return client.Set(ctx, buildTransactionKey(outTradeNo), string(tranBytes), 0).Err()
}

func GetTransaction(outTradeNo string) (model.Transaction, error) {
	var transaction model.Transaction
	tranStr, err := client.Get(ctx, buildTransactionKey(outTradeNo)).Result()
	if err != nil {
		return transaction, err
	}
	_ = json.Unmarshal([]byte(tranStr), &transaction)
	return transaction, err
}

func buildModeKey(user string) string {
	return "user:" + user + ":mode"
}

func GetMode(user string) (string, error) {
	result, err := client.Get(ctx, buildModeKey(user)).Result()
	if err == redis.Nil {
		return constant.GPT3, nil
	}
	return result, err
}

func SetMode(user string, mode string) error {
	return client.Set(ctx, buildModeKey(user), mode, 0).Err()
}

func buildErrorsKey(day string) string {
	return fmt.Sprintf("day:%s:errors", day)
}

func AppendError(day string, myErr model.MyError) error {
	errBytes, _ := json.Marshal(myErr)
	err := client.RPush(ctx, buildErrorsKey(day), string(errBytes)).Err()
	if err != nil {
		return err
	}
	err = client.Expire(ctx, buildErrorsKey(day), WEEK).Err()
	return err
}

func GetErrors(day string) ([]model.MyError, error) {
	var chatApiErrors []model.MyError
	errStrs, err := client.LRange(ctx, buildErrorsKey(day), 0, -1).Result()
	if err != nil {
		return chatApiErrors, err
	}

	for _, errStr := range errStrs {
		var chatApiError model.MyError
		_ = json.Unmarshal([]byte(errStr), &chatApiError)
		chatApiErrors = append(chatApiErrors, chatApiError)
	}
	return chatApiErrors, err
}

func GetErrorsLen(day string) (int64, error) {
	return client.LLen(ctx, buildErrorsKey(day)).Result()
}

func getInvitationCodeCursorKey() string {
	return "invitation-code-cursor"
}

func IncInvitationCodeCursor() (int64, error) {
	return client.Incr(ctx, getInvitationCodeCursorKey()).Result()
}

func buildInvitationCodeKey(user string) string {
	return "user:" + user + ":invitation-code"
}

func GetInvitationCode(user string) (string, error) {
	return client.Get(ctx, buildInvitationCodeKey(user)).Result()
}

func SetInvitationCode(user string, code string) error {
	return client.Set(ctx, buildInvitationCodeKey(user), code, 0).Err()
}

func buildUserKey(code string) string {
	return "invitation-code:" + code + ":user"
}

func GetUserByInvitationCode(code string) (string, error) {
	return client.Get(ctx, buildUserKey(code)).Result()
}

func SetUserByInvitationCode(code string, user string) error {
	return client.Set(ctx, buildUserKey(code), user, 0).Err()
}

func buildInviter(user string) string {
	return "user:" + user + ":inviter"
}

func SetInviter(user string, inviter string) error {
	return client.Set(ctx, buildInviter(user), inviter, 0).Err()
}

func GetInviter(user string) (string, error) {
	return client.Get(ctx, buildInviter(user)).Result()
}
