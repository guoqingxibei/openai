package handler

import (
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"time"
)

func onSubscribe(inMsg *wechat.Msg, writer http.ResponseWriter) {
	userName := inMsg.FromUserName
	log.Println("新增关注:", userName)
	_, err := gptredis.FetchSubscribeTimestamp(userName)
	if err == redis.Nil {
		err := gptredis.SetSubscribeTimestamp(userName, time.Now().Unix())
		if err != nil {
			log.Println("gptredis.SetSubscribeTimestamp failed", err)
		}
	}

	echoWechatTextMsg(writer, inMsg, constant.SubscribeReply)
}
