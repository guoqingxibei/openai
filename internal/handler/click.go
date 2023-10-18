package handler

import (
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/wechat"
)

func echoWechatOnClick(inMsg *wechat.Msg, writer http.ResponseWriter) {
	log.Printf("%s clicked the button 「%s」", inMsg.FromUserName, inMsg.EventKey)
	switch inMsg.EventKey {
	case constant.GPT3:
		fallthrough
	case constant.GPT4:
		switchGPTMode(inMsg.EventKey, inMsg, writer)
	case clear:
		clearHistory(inMsg, writer)
	case help:
		showUsage(inMsg, writer)
	case invite:
		getInvitationCode(inMsg, writer)
	case donate:
		fallthrough
	case group:
		fallthrough
	case contact:
		showImage(inMsg.EventKey, inMsg, writer)
	}
}
