package handler

import (
	"github.com/redis/go-redis/v9"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"openai/internal/constant"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

func onSubscribe(msg *message.MixMessage) *message.Reply {
	userName := string(msg.FromUserName)
	log.Println("新增关注:", userName)
	_, err := store.GetSubscribeTimestamp(userName)
	if err == redis.Nil {
		err := store.SetSubscribeTimestamp(userName, time.Now().Unix())
		if err != nil {
			log.Println("store.SetSubscribeTimestamp failed", err)
		}
	}
	return util.BuildTextReply(constant.SubscribeReply)
}
