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
		return fmt.Sprintf("已切换到「midjourney画图」模式，每次画图消耗次数%d。"+
			"你的付费次数剩余%d次，可以<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。"+
			"\n\n在此模式下，你用英文给出描述(prompt)，稍等片刻，公众号返回midjourney生成的4张作品。"+
			"在使用此模式前，请确保阅读过<a href=\"%s\">这篇教程</a>。",
			logic.GetTimesPerQuestion(mode),
			balance,
			util.GetPayLink(userName),
			util.GetInvitationTutorialLink(),
			"https://cxyds.top/2023/10/27/midjourney.html",
		)
	}

	return fmt.Sprintf("已切换到「%s」模式，每次对话消耗次数%d。"+
		"你的付费次数剩余%d次，可以<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。",
		mode,
		logic.GetTimesPerQuestion(mode),
		balance,
		util.GetPayLink(userName),
		util.GetInvitationTutorialLink(),
	)
}
