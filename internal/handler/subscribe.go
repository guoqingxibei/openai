package handler

import (
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/wechat"
	"openai/internal/store"
	"time"
)

func onSubscribe(inMsg *wechat.Msg, writer http.ResponseWriter) {
	userName := inMsg.FromUserName
	log.Println("新增关注:", userName)
	_, err := store.GetSubscribeTimestamp(userName)
	if err == redis.Nil {
		err := store.SetSubscribeTimestamp(userName, time.Now().Unix())
		if err != nil {
			log.Println("store.SetSubscribeTimestamp failed", err)
		}
	}

	echoWechatTextMsg(writer, inMsg, constant.SubscribeReply)
}
