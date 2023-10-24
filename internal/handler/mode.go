package handler

import (
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/store"
	"openai/internal/util"
)

func switchMode(mode string, msg *message.MixMessage) *message.Reply {
	userName := string(msg.FromUserName)
	_ = store.SetMode(userName, mode)
	return util.BuildTextReply(buildModeDesc(userName, mode))
}

func buildModeDesc(userName string, mode string) (desc string) {
	desc = fmt.Sprintf("已切换到「%s」模式，每次对话消耗次数%d。", mode, logic.GetTimesPerQuestion(mode))
	if mode == constant.GPT3 {
		return
	}

	balance, _ := store.GetPaidBalance(userName)
	desc += fmt.Sprintf("付费次数剩余%d次，可以<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。",
		balance,
		util.GetPayLink(userName),
		util.GetInvitationTutorialLink(),
	)
	return
}
