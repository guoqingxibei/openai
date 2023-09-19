package handler

import (
	"fmt"
	"net/http"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"openai/internal/util"
)

func switchGPTMode(gptMode string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	userName := inMsg.FromUserName
	_ = gptredis.SetGPTMode(userName, gptMode)
	echoWechatTextMsg(writer, inMsg, buildModeDesc(userName, gptMode))
}

func buildModeDesc(userName string, gptMode string) string {
	desc := fmt.Sprintf("已切换到「%s」模式，每次提问消耗次数%d。", gptMode, logic.GetTimesPerQuestion(gptMode))
	balance, _ := gptredis.FetchPaidBalance(userName)
	if gptMode == constant.GPT4 {
		desc += fmt.Sprintf("付费次数剩余%d次，<a href=\"%s\">点我购买次数</a>。", balance, util.GetPayLink(userName))
	}
	return desc
}
