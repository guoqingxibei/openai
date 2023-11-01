package handler

import (
	"errors"
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/service/wechat"
	"openai/internal/util"
	"runtime/debug"
)

type ChatRound struct {
	question string
	answer   string
}

func ServeWechat(rw http.ResponseWriter, req *http.Request) {
	officialAccount := wechat.GetAccount()

	// 传入request和responseWriter
	server := officialAccount.GetServer(req, rw)
	server.SetParseXmlToMsgFn(util.ParseXmlToMsg)

	//设置接收消息的处理方法
	server.SetMessageHandler(talk)

	//处理消息接收以及回复
	err := server.Serve()
	if err != nil {
		errorx.RecordError("server.Serve() failed", err)
		err = server.BuildResponse(util.BuildTextReply(constant.TryAgain))
		if err != nil {
			errorx.RecordError("server.BuildResponse() failed", err)
			return
		}
	}

	//发送回复的消息
	err = server.Send()
	if err != nil {
		errorx.RecordError("server.Send() failed", err)
	}
}

// Talk https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Passive_user_reply_message.html
// 微信服务器在五秒内收不到响应会断掉连接，并且重新发起请求，总共重试三次
func talk(msg *message.MixMessage) (reply *message.Reply) {
	defer func() {
		if r := recover(); r != nil {
			panicMsg := fmt.Sprintf("%v\n%s", r, debug.Stack())
			errorx.RecordError("panic captured", errors.New(panicMsg))
			reply = util.BuildTextReply(constant.TryAgain)
		}
	}()

	var err error
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
		reply, err = onReceiveText(msg)
	default:
		log.Printf("未实现的消息类型: %s\n", msg.MsgType)
		reply = util.BuildTextReply("抱歉，目前还只支持文本和语音消息哦~")
	}
	if err != nil {
		errorx.RecordError("Talk() failed", err)
		return util.BuildTextReply(constant.TryAgain)
	}
	return reply
}
