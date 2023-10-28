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
	balance, _ := store.GetPaidBalance(userName)
	if mode == constant.Draw {
		return fmt.Sprintf("已切换到「%s」模式，每次画图消耗次数%d。"+
			"你的付费额度剩余%d次，<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。"+
			"\n\n此模式下，你需要用英文给出描述，稍后midjourney将为你奉上精美作品。"+
			"开始之前，请务必仔细阅读<a href=\"%s\">这篇教程</a>。",
			logic.GetModeName(mode),
			logic.GetTimesPerQuestion(mode),
			balance,
			util.GetPayLink(userName),
			util.GetInvitationTutorialLink(),
			"https://cxyds.top/2023/10/27/midjourney.html",
		)
	}

	return fmt.Sprintf("已切换到「%s」模式，每次对话消耗次数%d。"+
		"你的付费额度剩余%d次，<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。",
		logic.GetModeName(mode),
		logic.GetTimesPerQuestion(mode),
		balance,
		util.GetPayLink(userName),
		util.GetInvitationTutorialLink(),
	)
}
