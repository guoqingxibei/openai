package handler

import (
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
	report  = "report"
)

var keywords = [5]string{donate, help, contact, report}

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
	case report:
		showReport(inMsg, writer)
	}
	return true
}

func showContactInfo(inMsg *wechat.Msg, writer http.ResponseWriter) {
	echoWechatTextMsg(writer, inMsg, constant.ContactInfo)
}

func showReport(inMsg *wechat.Msg, writer http.ResponseWriter) {
	echoWechatTextMsg(writer, inMsg, constant.ReportInfo)
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
	usage := logic.BuildChatUsage(userName)
	usage += "\n\n" + constant.ContactDesc + "\n" + constant.DonateDesc
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
