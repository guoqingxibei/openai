package store

import (
	"context"
	"github.com/redis/go-redis/v9"
	"openai/internal/config"
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

func IncRequestTimesForMsg(msgId int64) (int64, error) {
	msgIdStr := strconv.FormatInt(msgId, 10)
	key := buildRequestTimesKey(msgIdStr)
	times, err := client.Incr(ctx, key).Result()
	if times == 1 {
		client.Expire(ctx, key, time.Second*30)
	}
	if err != nil {
		return 0, nil
	}
	return times, nil
}

func buildRequestTimesKey(msgIdStr string) string {
	return "msg-id:" + msgIdStr + ":request-times"
}

func buildUsedTimesKey(user string) string {
	return "user:" + user + ":used-times"
}

func IncUsedTimes(user string) (int, error) {
	times, err := client.Incr(ctx, buildUsedTimesKey(user)).Result()
	return int(times), err
}
