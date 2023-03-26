package handler

import (
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
)

const (
	Chat  = "chat"
	Image = "image"
)

var modes = [2]string{Chat, Image}

func checkModeSwitch(question string) (string, bool) {
	for _, mode := range modes {
		if question == "/"+mode {
			return mode, true
		}
	}
	return "", false
}

func setModeThenReply(mode string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	err := gptredis.SetModeForUser(inMsg.FromUserName, mode)
	if err != nil {
		log.Println("gptredis.SetModeForUser failed", err)
		echoWechatTextMsg(writer, inMsg, constant.TryAgain)
	} else {
		echoWechatTextMsg(writer, inMsg, buildReplyForMode(mode))
	}
}

func buildReplyForMode(mode string) string {
	reply := "已切换到 " + mode + " 模式"
	if mode == Image {
		reply += "\n用法：你说一句文字描述，我回一张图片，单轮对话。"
	}
	return reply
}
