package handler

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"strings"
)

const (
	donate  = "donate"
	help    = "help"
	contact = "contact"
)

var keywords = [5]string{donate, help, contact, constant.Chat, constant.Image}

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
	case contact:
		showContactInfo(inMsg, writer)
	case donate:
		showDonateQr(inMsg, writer)
	case help:
		showUsage(inMsg, writer)
	case constant.Chat:
		fallthrough
	case constant.Image:
		switchMode(keyword, inMsg, writer)
	}
	return true
}

func showContactInfo(inMsg *wechat.Msg, writer http.ResponseWriter) {
	echoWechatTextMsg(writer, inMsg, constant.ContactInfo)
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
	userName := inMsg.FromUserName
	mode, err := gptredis.FetchModeForUser(userName)
	if err != nil {
		if err != redis.Nil {
			log.Println("gptredis.FetchModeForUser failed", err)
			echoWechatTextMsg(writer, inMsg, constant.TryAgain)
			return
		}
		mode = constant.Chat
	}
	usage := fmt.Sprintf("当前是%s模式。", mode)
	usage += "\n\n回复chat，开启chat模式，" + logic.BuildChatUsage(userName)
	usage += "\n\n回复image，开启image模式，" + logic.BuildImageUsage(userName)
	usage += "\n\n" + constant.ContactDesc
	echoWechatTextMsg(writer, inMsg, usage)
}

func switchMode(mode string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	userName := inMsg.FromUserName
	err := gptredis.SetModeForUser(userName, mode)
	if err != nil {
		log.Println("gptredis.SetModeForUser failed", err)
		echoWechatTextMsg(writer, inMsg, constant.TryAgain)
	} else {
		echoWechatTextMsg(writer, inMsg, buildReplyWhenSwitchMode(userName, mode))
	}
}

func buildReplyWhenSwitchMode(userName string, mode string) string {
	reply := "已切换到" + mode + "模式，"
	if mode == constant.Image {
		reply += logic.BuildImageUsage(userName)
	} else {
		reply += logic.BuildChatUsage(userName)
	}
	return reply
}
