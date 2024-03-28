package handler

import (
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

func onSubscribe(msg *message.MixMessage) *message.Reply {
	userName := string(msg.FromUserName)
	log.Println("新增关注:", userName)
	_, err := store.GetSubscribeTimestamp(userName)
	if errors.Is(err, redis.Nil) {
		_ = store.SetSubscribeTimestamp(userName, time.Now().Unix())
	}

	var reply string
	if util.AccountIsUncle() {
		reply = "此公众号已接入ChatGPT 3.5、4，<a href=\"https://cxyds.top/2023/07/03/faq.html\">点我了解详细用法</a>。" +
			"\n\n现在，你可以直接用文字或语音和我对话咯~"
	} else {
		reply = "此公众号已接入ChatGPT 3.5、4和MidJourney，<a href=\"https://cxyds.top/2023/07/03/faq.html\">点我了解详细用法</a>。" +
			"\n\n现在，你可以直接用文字或语音和我对话咯~"
	}
	return util.BuildTextReply(reply)
}
