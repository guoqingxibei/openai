package handler

import (
	"fmt"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
)

func switchGPTMode(keyword string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	_ = gptredis.SetGPTMode(inMsg.FromUserName, keyword)
	echoWechatTextMsg(writer, inMsg, buildModeDesc(keyword))
}

func buildModeDesc(keyword string) string {
	desc := fmt.Sprintf("已切换到「%s」模式，", keyword)
	if keyword == constant.GPT3 {
		desc += "每次提问消耗次数1。"
	} else {
		desc += "每次提问消耗次数10。"
	}
	return desc
}
