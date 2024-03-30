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
	switch mode {
	case constant.GPT3:
		fallthrough
	case constant.GPT4:
		desc = fmt.Sprintf("已切换到「%s」模式，每次对话消耗次数%d。"+
			"你的付费额度剩余%d次，<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。",
			logic.GetModeName(mode),
			logic.GetTimesPerQuestion(mode),
			balance,
			util.GetPayLink(userName),
			util.GetInvitationTutorialLink(),
		)
	case constant.Draw:
		desc = fmt.Sprintf("已切换到「%s」模式，每次绘画消耗次数%d。"+
			"你的付费额度剩余%d次，<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。"+
			"\n\n在此模式下，你给出图片描述，稍后midjourney为你奉上精美作品。"+
			"开始之前，请务必仔细阅读<a href=\"%s\">这篇教程</a>。",
			logic.GetModeName(mode),
			logic.GetTimesPerQuestion(mode),
			balance,
			util.GetPayLink(userName),
			util.GetInvitationTutorialLink(),
			"https://cxyds.top/2023/10/27/ai-draw.html",
		)
	case constant.TTS:
		desc = fmt.Sprintf("已切换到「%s」模式，每%d字消耗次数1。"+
			"你的付费额度剩余%d次，<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。"+
			"\n\n在此模式下，你输入文字，OpenAI为你转换成语音。",
			logic.GetModeName(mode),
			constant.CharCountPerTimeTTS,
			balance,
			util.GetPayLink(userName),
			util.GetInvitationTutorialLink(),
		)
	case constant.Translate:
		desc = fmt.Sprintf("已切换到「%s」模式，每次翻译消耗次数%d，此模式由OpenAI提供技术支持。"+
			"你的付费额度剩余%d次，<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。",
			logic.GetModeName(mode),
			logic.GetTimesPerQuestion(mode),
			balance,
			util.GetPayLink(userName),
			util.GetInvitationTutorialLink(),
		)
	}
	return
}
