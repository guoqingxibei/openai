package handler

import (
	"log"
	"net/http"
	"openai/internal/service/wechat"
)

func echoWechatOnClick(inMsg *wechat.Msg, writer http.ResponseWriter) {
	log.Printf("%s clicked the button 「%s」", inMsg.FromUserName, inMsg.EventKey)
	switch inMsg.EventKey {
	case help:
		showUsage(inMsg, writer)
	case donate:
		fallthrough
	case group:
		fallthrough
	case contact:
		showImage(inMsg.EventKey, inMsg, writer)
	}
}
