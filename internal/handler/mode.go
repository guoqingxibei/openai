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

func switchMode(mode string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	userName := inMsg.FromUserName
	_ = gptredis.SetMode(userName, mode)
	echoWechatTextMsg(writer, inMsg, buildModeDesc(userName, mode))
}

func buildModeDesc(userName string, mode string) string {
	desc := fmt.Sprintf("已切换到「%s」模式，每次提问消耗次数%d。", mode, logic.GetTimesPerQuestion(mode))
	balance, _ := gptredis.FetchPaidBalance(userName)
	if mode == constant.GPT4 {
		desc += fmt.Sprintf("付费次数剩余%d次，可<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。",
			balance,
			util.GetPayLink(userName),
			util.GetInvitationTutorialLink(),
		)
	}
	return desc
}
