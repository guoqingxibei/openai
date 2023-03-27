package handler

import (
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"strings"
)

// mode
const (
	Chat  = "chat"
	Image = "image"
)

var modes = [2]string{Chat, Image}

func checkModeSwitch(question string) (string, bool) {
	question = strings.ToLower(question)
	for _, mode := range modes {
		if question == mode {
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
	reply := "已切换到 " + mode + " 模式，"
	if mode == Image {
		reply += "你说一句尽可能完整的图片描述，我画一张对应的图片，单轮对话，每天仅限 5 次（成本昂贵，敬请谅解）。"
	} else {
		reply += "你问我答，多轮对话，不限次数。"
	}
	return reply + "\n\n回复 help，可查看详细用法。\n回复 donate，可捐赠作者。"
}
