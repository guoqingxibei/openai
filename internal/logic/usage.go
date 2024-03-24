package logic

import (
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"openai/internal/store"
	"openai/internal/util"
)

func ShowUsage(msg *message.MixMessage) (reply *message.Reply) {
	user := string(msg.FromUserName)
	mode, _ := store.GetMode(user)
	usage := "【模式】" + GetModeDesc(mode)

	usage += "\n【额度】"
	paidBalance, err := store.GetPaidBalance(user)
	hasPaid := !errors.Is(err, redis.Nil)
	if !hasPaid {
		usage += fmt.Sprintf("免费额度剩余%d次，每天免费%d次。", GetBalance(user), GetQuota(user))
	}
	usage += fmt.Sprintf("付费额度剩余%d次，<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。",
		paidBalance,
		util.GetPayLink(user),
		util.GetInvitationTutorialLink(),
	)
	usage += "\n\n<a href=\"https://cxyds.top/2023/07/03/faq.html\">更多用法</a>" +
		" | <a href=\"https://cxyds.top/2023/09/17/group-qr.html\">交流群</a>" +
		" | <a href=\"https://cxyds.top/2023/09/17/writer-qr.html\">联系作者</a>"
	return util.BuildTextReply(usage)
}
