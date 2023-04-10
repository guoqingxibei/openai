package handler

import (
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"time"
)

func onSubscribe(inMsg *wechat.Msg, writer http.ResponseWriter) {
	log.Println("新增关注:", inMsg.FromUserName)
	err := gptredis.SetSubscribeTimestamp(inMsg.FromUserName, time.Now().Unix())
	if err != nil {
		log.Println("gptredis.SetSubscribeTimestamp failed", err)
	}
	echoWechatTextMsg(writer, inMsg, constant.SubscribeReply)
}
