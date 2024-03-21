package logic

import (
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"openai/internal/store"
	"openai/internal/util"
)

func ShowUsage(msg *message.MixMessage) (reply *message.Reply) {
	user := string(msg.FromUserName)
	mode, _ := store.GetMode(user)
	usage := "【模式】" + GetModeDesc(mode)

	usage += "\n【额度】"
	paidBalance, _ := store.GetPaidBalance(user)
	if paidBalance <= 0 {
		usage += fmt.Sprintf("免费额度剩余%d次，每天免费%d次。", GetBalance(user), GetQuota(user))
	}
	usage += fmt.Sprintf("付费额度剩余%d次，<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。",
		paidBalance,
		util.GetPayLink(user),
		util.GetInvitationTutorialLink(),
	)
	usage += "\n\n<a href=\"https://cxyds.top/2023/07/03/faq.html\">更多用法</a>" +
		" | <a href=\"https://cxyds.top/2023/09/17/group-qr.html\">交流群</a>" +
		" | <a href=\"https://cxyds.top/2023/09/17/writer-qr.html\">联系作者</a>"
	return util.BuildTextReply(usage)
}
