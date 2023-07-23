package handler

import (
	"encoding/json"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"log"
	"net/http"
	"openai/internal/logic"
	"openai/internal/service/gptredis"
	wechatService "openai/internal/service/wechat"
	"time"
)

type transactionReq struct {
	OpenId      string `json:"openid"`
	PriceInFen  int    `json:"price_in_fen"`
	Times       int    `json:"times"`
	Description string `json:"description"`
}

func Transaction(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var transactionReq transactionReq
	err := decoder.Decode(&transactionReq)
	if err != nil {
		log.Println(err)
		return
	}

	prepayId, err := logic.InitiateTransaction(
		transactionReq.OpenId, transactionReq.PriceInFen, transactionReq.Times, transactionReq.Description,
	)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	params, err := wechatService.GeneratePaySignParams(prepayId)
	if err != nil {
		log.Println(err)
		return
	}
	data, _ := json.Marshal(params)
	w.Write(data)
}

func NotifyTransactionResult(w http.ResponseWriter, r *http.Request) {
	notifyReq, err := wechat.V3ParseNotify(r)
	if err != nil {
		log.Println(err)
		return
	}

	result, err := wechatService.VerifySignAndDecrypt(notifyReq)
	if err != nil {
		log.Println(err)
		return
	}

	outTradeNo := result.OutTradeNo
	transaction, _ := gptredis.FetchTransaction(outTradeNo)
	transaction.TradeState = result.TradeState
	payload, _ := json.Marshal(result)
	transaction.Payload = string(payload)
	transaction.UpdatedTime = time.Now().Unix()

	if result.TradeState == "SUCCESS" {
		if !transaction.Redeemed {
			openId := transaction.OpenId
			times := transaction.Times
			balance, _ := gptredis.FetchPaidBalance(openId)
			_ = gptredis.SetPaidBalance(openId, times+balance)
			transaction.Redeemed = true
		}
	}
	_ = gptredis.SetTransaction(outTradeNo, transaction)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(wechat.V3NotifyRsp{Code: gopay.SUCCESS, Message: "成功"})
	w.Write(data)
}
