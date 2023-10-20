package handler

import (
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"openai/internal/constant"
	"openai/internal/util"
	"runtime/debug"
)

type ChatRound struct {
	question string
	answer   string
}

// Talk https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Passive_user_reply_message.html
// 微信服务器在五秒内收不到响应会断掉连接，并且重新发起请求，总共重试三次
func Talk(msg *message.MixMessage) (reply *message.Reply) {
	// unhandled exception
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Captured panic:", r, string(debug.Stack()))
			reply = util.BuildTextReply(constant.TryAgain)
		}
	}()

	// 非文本不回复(返回success表示不回复)
	switch msg.MsgType {
	case message.MsgTypeEvent:
		switch msg.Event {
		case message.EventSubscribe:
			reply = onSubscribe(msg)
		case message.EventUnsubscribe:
			log.Println("取消关注:", msg.FromUserName)
			reply = util.BuildTextReply("")
		case message.EventClick:
			reply = onClick(msg)
		default:
			log.Printf("未实现的事件: %s\n", msg.Event)
			reply = util.BuildTextReply("")
		}
	case message.MsgTypeVoice:
		fallthrough
	case message.MsgTypeText:
		reply = onReceiveText(msg)
	default:
		log.Printf("未实现的消息类型: %s\n", msg.MsgType)
		reply = util.BuildTextReply("目前还只支持文本和语音消息哦~")
	}
	return reply
}
