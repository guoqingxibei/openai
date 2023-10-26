package store

import (
	"encoding/json"
	"fmt"
	"openai/internal/model"
	"time"
)

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

func buildOpenIdKey(authCode string) string {
	return "auth-code:" + authCode + ":open-id"
}

func GetOpenId(authCode string) (string, error) {
	return client.Get(ctx, buildOpenIdKey(authCode)).Result()
}

func SetOpenId(authCode string, openId string) error {
	return client.Set(ctx, buildOpenIdKey(authCode), openId, time.Hour*12).Err()
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
