package handler

import (
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/wechat"
)

const (
	donate = "donate"
	help   = "help"
)

var keywords = [2]string{donate, help}

func hitKeyword(inMsg *wechat.Msg, writer http.ResponseWriter) bool {
	question := inMsg.Content
	var keyword string
	for _, word := range keywords {
		if question == word {
			keyword = word
			break
		}
	}
	if keyword == "" {
		return false
	}

	switch keyword {
	case donate:
		QrMediaId, err := wechat.GetMediaIdOfDonateQr()
		if err != nil {
			log.Println("wechat.GetMediaIdOfDonateQr failed", err)
			echoWechatTextMsg(writer, inMsg, constant.TryAgain)
			return true
		}
		echoWechatImageMsg(writer, inMsg, QrMediaId)
	case help:
		showUsage(inMsg, writer)
	}
	return true
}
