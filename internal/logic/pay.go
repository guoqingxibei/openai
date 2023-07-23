package logic

import (
	"github.com/go-pay/gopay/pkg/util"
	"log"
	"openai/internal/models"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"time"
)

func InitiateTransaction(openid string, price int, times int, description string) (string, error) {
	tradeNo := util.RandomString(32)
	log.Println("tradeNo:", tradeNo)
	prepayId, err := wechat.InitiateTransaction(openid, tradeNo, price*100, description)
	if err != nil {
		return "", err
	}

	now := time.Now().Unix()
	transaction := models.Transaction{
		OutTradeNo:  tradeNo,
		OpenId:      openid,
		PrepayId:    prepayId,
		TradeState:  "created",
		Price:       price,
		Times:       times,
		Description: description,
		CreatedTime: now,
		UpdatedTime: now,
	}
	_ = gptredis.SetTransaction(tradeNo, transaction)
	return prepayId, err
}
