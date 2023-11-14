package store

import (
	"encoding/json"
	"fmt"
	"openai/internal/model"
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

func buildSuccessOutTradeNosKey(day string) string {
	return fmt.Sprintf("day:%s:success-out-trade-nos", day)
}

func AppendSuccessOutTradeNo(day string, outTradeNo string) (err error) {
	err = client.RPush(ctx, buildSuccessOutTradeNosKey(day), outTradeNo).Err()
	if err != nil {
		return
	}

	err = client.Expire(ctx, buildSuccessOutTradeNosKey(day), WEEK).Err()
	return
}

func GetSuccessOutTradeNos(day string) ([]string, error) {
	return client.LRange(ctx, buildSuccessOutTradeNosKey(day), 0, -1).Result()
}
