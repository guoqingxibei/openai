package handler

import (
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"strings"
)

const (
	donate = "donate"
	help   = "help"
)

// mode
const (
	Chat  = "chat"
	Image = "image"
)

var keywords = [4]string{donate, help, Chat, Image}

func hitKeyword(inMsg *wechat.Msg, writer http.ResponseWriter) bool {
	question := inMsg.Content
	question = strings.TrimSpace(question)
	question = strings.ToLower(question)
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
		showDonateQr(inMsg, writer)
	case help:
		showUsage(inMsg, writer)
	case Chat:
		fallthrough
	case Image:
		switchMode(keyword, inMsg, writer)
	}
	return true
}

func showDonateQr(inMsg *wechat.Msg, writer http.ResponseWriter) {
	QrMediaId, err := wechat.GetMediaIdOfDonateQr()
	if err != nil {
		log.Println("wechat.GetMediaIdOfDonateQr failed", err)
		echoWechatTextMsg(writer, inMsg, constant.TryAgain)
		return
	}
	echoWechatImageMsg(writer, inMsg, QrMediaId)
}

func showUsage(inMsg *wechat.Msg, writer http.ResponseWriter) {
	mode, err := gptredis.FetchModeForUser(inMsg.FromUserName)
	if err != nil {
		if err != redis.Nil {
			log.Println("gptredis.FetchModeForUser failed", err)
			echoWechatTextMsg(writer, inMsg, constant.TryAgain)
			return
		}
		mode = Chat
	}
	usage := "当前是 " + mode + " 模式。"
	usage += "\n\n回复 chat，开启 chat 模式。此模式是默认模式，在此模式下，" + constant.ChatUsage
	usage += "\n\n回复 image，开启 image 模式。在此模式下，" + constant.ImageUsage
	usage += "\n\n" + constant.DonateDesc
	echoWechatTextMsg(writer, inMsg, usage)
}

func switchMode(mode string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	err := gptredis.SetModeForUser(inMsg.FromUserName, mode)
	if err != nil {
		log.Println("gptredis.SetModeForUser failed", err)
		echoWechatTextMsg(writer, inMsg, constant.TryAgain)
	} else {
		echoWechatTextMsg(writer, inMsg, buildReplyWhenSwitchMode(mode))
	}
}

func buildReplyWhenSwitchMode(mode string) string {
	reply := "已切换到 " + mode + " 模式，"
	if mode == Image {
		reply += constant.ImageUsage
	} else {
		reply += constant.ChatUsage
	}
	return reply + "\n\n" + constant.UsageTail
}
