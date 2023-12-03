package handler

import (
	"encoding/json"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"log"
	"net/http"
	"openai/internal/logic"
	"openai/internal/service/errorx"
	wechatService "openai/internal/service/wechat"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

type transactionReq struct {
	OpenId      string `json:"openid"`
	UncleOpenId string `json:"uncle_openid"`
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
		errorx.RecordError("decoder.Decode() failed", err)
		return
	}

	prepayId, outTradeNo, err := logic.InitiateTransaction(
		transactionReq.OpenId,
		transactionReq.UncleOpenId,
		transactionReq.PriceInFen,
		transactionReq.Times,
		transactionReq.Description,
	)
	if err != nil {
		errorx.RecordError("logic.InitiateTransaction() failed", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	params, err := wechatService.GeneratePaySignParams(prepayId)
	if err != nil {
		errorx.RecordError("wechatService.GeneratePaySignParams() failed", err)
		return
	}
	data, _ := json.Marshal(transactionRes{params, outTradeNo})
	w.Write(data)
}

func NotifyTransactionResult(w http.ResponseWriter, r *http.Request) {
	notifyReq, err := wechat.V3ParseNotify(r)
	if err != nil {
		errorx.RecordError("wechat.V3ParseNotify() failed", err)
		fail(w)
		return
	}

	result, err := wechatService.VerifySignAndDecrypt(notifyReq)
	if err != nil {
		log.Println("wechatService.VerifySignAndDecrypt() failed", err)
		fail(w)
		return
	}

	outTradeNo := result.OutTradeNo
	transaction, _ := store.GetTransaction(outTradeNo)
	transaction.TradeState = result.TradeState
	payload, _ := json.Marshal(result)
	transaction.Payload = string(payload)
	transaction.UpdatedAt = time.Now()

	if result.TradeState == "SUCCESS" {
		if !transaction.Redeemed {
			_ = store.AppendSuccessOutTradeNo(util.Today(), outTradeNo)
			openId := transaction.OpenId
			times := transaction.Times
			useUncleDB := false
			if transaction.UncleOpenId != "" {
				openId = transaction.UncleOpenId
				useUncleDB = true
			}
			balance, _ := store.GetPaidBalanceWithDB(openId, useUncleDB)
			_ = store.SetPaidBalanceWithDB(openId, times+balance, useUncleDB)
			transaction.Redeemed = true
		}
	}
	_ = store.SetTransaction(outTradeNo, transaction)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(wechat.V3NotifyRsp{Code: gopay.SUCCESS, Message: "成功"})
	w.Write(data)
}

func GetTradeResult(w http.ResponseWriter, r *http.Request) {
	outTradeId := r.URL.Query().Get("out_trade_no")
	transaction, err := store.GetTransaction(outTradeId)
	if err != nil {
		errorx.RecordError("store.GetTransaction() failed", err)
		return
	}

	openId := transaction.OpenId
	useUncleDB := false
	if transaction.UncleOpenId != "" {
		openId = transaction.UncleOpenId
		useUncleDB = true
	}
	balance, _ := store.GetPaidBalanceWithDB(openId, useUncleDB)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(tradeResult{balance, transaction.Redeemed})
	w.Write(data)
}

func fail(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	data, _ := json.Marshal(wechat.V3NotifyRsp{Code: gopay.FAIL, Message: "失败"})
	w.Write(data)
}
