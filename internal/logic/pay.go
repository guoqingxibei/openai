package logic

import (
	"github.com/go-pay/util"
	"log/slog"
	"openai/internal/model"
	"openai/internal/service/wechat"
	"openai/internal/store"
	"time"
)

func InitiateTransaction(
	openid string,
	uncleOpenid string,
	priceInFen int,
	times int,
	description string) (string, string, error) {
	tradeNo := util.RandomString(32)
	slog.Info("InitiateTransaction", "tradeNo", tradeNo)
	prepayId, err := wechat.InitiateTransaction(openid, tradeNo, priceInFen, description)
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	transaction := model.Transaction{
		OutTradeNo:  tradeNo,
		OpenId:      openid,
		UncleOpenId: uncleOpenid,
		PrepayId:    prepayId,
		TradeState:  "",
		Redeemed:    false,
		PriceInFen:  priceInFen,
		Times:       times,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_ = store.SetTransaction(tradeNo, transaction)
	return prepayId, tradeNo, err
}

func IsAround8PM() bool {
	now := time.Now()

	today := now.Format("2006-01-02")
	startTime, _ := time.Parse("2006-01-02 15:04", today+" 20:00")
	endTime, _ := time.Parse("2006-01-02 15:04", today+" 20:05")

	return now.After(startTime) && now.Before(endTime)
}
