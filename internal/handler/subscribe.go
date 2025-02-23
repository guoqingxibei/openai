package handler

import (
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log/slog"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

func onSubscribe(msg *message.MixMessage) *message.Reply {
	userName := string(msg.FromUserName)
	slog.Info("新增关注:" + userName)
	_, err := store.GetSubscribeTimestamp(userName)
	if errors.Is(err, redis.Nil) {
		_ = store.SetSubscribeTimestamp(userName, time.Now().Unix())
	}

	var reply string
	if util.AccountIsUncle() {
		reply = "此公众号已接入ChatGPT和DeepSeek，<a href=\"https://cxyds.top/2023/07/03/faq.html\">点我了解详细用法</a>。" +
			"\n\n现在，你可以直接用文字、图片和我对话咯~"
	} else {
		reply = "此公众号已接入ChatGPT、DeepSeek和MidJourney，<a href=\"https://cxyds.top/2023/07/03/faq.html\">点我了解详细用法</a>。" +
			"\n\n现在，你可以直接用文字、图片和我对话咯~"
	}
	return util.BuildTextReply(reply)
}
