package handler

import (
	"io"
	"log"
	"net/http"
	"openai/internal/service/wechat"
)

var (
	success = []byte("success")
)

type ChatRound struct {
	question string
	answer   string
}

// Talk https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Passive_user_reply_message.html
// 微信服务器在五秒内收不到响应会断掉连接，并且重新发起请求，总共重试三次
func Talk(writer http.ResponseWriter, request *http.Request) {
	bs, _ := io.ReadAll(request.Body)
	inMsg := wechat.NewInMsg(bs)

	if inMsg == nil {
		echoWeChat(writer, []byte("xml格式公众号消息接口，请勿手动调用"))
		return
	}

	// 非文本不回复(返回success表示不回复)
	switch inMsg.MsgType {
	case "event":
		switch inMsg.Event {
		case "subscribe":
			onSubscribe(inMsg, writer)
		case "unsubscribe":
			log.Println("取消关注:", inMsg.FromUserName)
			echoWeChat(writer, success)
		case "CLICK":
			echoWechatOnClick(inMsg, writer)
		default:
			log.Printf("未实现的事件: %s\n", inMsg.Event)
			echoWeChat(writer, success)
		}
	case "voice":
		fallthrough
	case "text":
		echoText(inMsg, writer)
	default:
		log.Printf("未实现的消息类型: %s\n", inMsg.MsgType)
		echoWechatTextMsg(writer, inMsg, "目前还只支持文本和语音消息哦~")
	}
}

func echoWeChat(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func echoWechatTextMsg(writer http.ResponseWriter, inMsg *wechat.Msg, reply string) {
	outMsg := inMsg.BuildTextMsg(reply)
	echoWeChat(writer, outMsg)
}
