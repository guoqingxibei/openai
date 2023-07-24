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

type transactionRes struct {
	Params     *wechat.JSAPIPayParams `json:"params"`
	OutTradeNo string                 `json:"out_trade_no"`
}

type tradeResult struct {
	PaidBalance int  `json:"paid_balance"`
	Redeemed    bool `json:"redeemed"`
}

func Transaction(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var transactionReq transactionReq
	err := decoder.Decode(&transactionReq)
	if err != nil {
		log.Println(err)
		return
	}

	prepayId, outTradeNo, err := logic.InitiateTransaction(
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
	data, _ := json.Marshal(transactionRes{params, outTradeNo})
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

func GetTradeResult(w http.ResponseWriter, r *http.Request) {
	outTradeId := r.URL.Query().Get("out_trade_no")
	transaction, err := gptredis.FetchTransaction(outTradeId)
	if err != nil {
		log.Println(err)
		return
	}
	balance, _ := gptredis.FetchPaidBalance(transaction.OpenId)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(tradeResult{balance, transaction.Redeemed})
	w.Write(data)
}