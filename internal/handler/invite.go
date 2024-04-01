package handler

import (
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"openai/internal/logic"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

const (
	inviterReward = 15
)
const sizeOfCode = 6 // the length of invitation code
const halfAnHour = 30 * 60
const inviteTutorial = `【邀请码】
%s

【流程】
1 分享公众号给你的朋友关注
2 让ta向公众号发送此邀请码

【奖励】
%d次的额度/邀请

<a href="%s">「查看详情」</a>`
const inviteSuccessMsg = `【成功接受邀请】系统已为你的邀请者充值%d次的额度，快去告诉ta吧！

<a href="%s">「如何邀请好友」</a>`

var codeChars = []rune{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K',
				'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
var base = len(codeChars) // char count

func getInvitationCode(msg *message.MixMessage) (reply *message.Reply) {
	user := string(msg.FromUserName)
	code, _ := store.GetInvitationCode(user)
	if code == "" {
		cursor, _ := store.IncInvitationCodeCursor()
		code = convertToInvitationCode(int(cursor - 1))
		_ = store.SetInvitationCode(user, code)
		_ = store.SetUserByInvitationCode(code, user)
	}
	return util.BuildTextReply(fmt.Sprintf(inviteTutorial,
		code,
		inviterReward,
		util.GetInvitationTutorialLink(),
	))
}

func convertToInvitationCode(n int) string {
	baseArr := make([]int, sizeOfCode)
	for i := sizeOfCode - 1; i >= 0; i-- {
		baseArr[i] = n % base
		n = n / base
	}

	code := ""
	for i := 0; i < len(baseArr); i++ {
		code += string(codeChars[baseArr[i]])
	}
	return code
}

func doInvite(inviter string, msg *message.MixMessage) (reply *message.Reply) {
	user := string(msg.FromUserName)
	if user == inviter {
		return util.BuildTextReply("抱歉，你无法使用自己的邀请码。")
	}

	currentTimestamp := time.Now().Unix()
	subScribeTimestamp, _ := store.GetSubscribeTimestamp(user)
	if currentTimestamp-subScribeTimestamp > halfAnHour {
		return util.BuildTextReply("抱歉，邀请码仅在首次关注公众号半小时内输入有效。")
	}

	existedInviter, _ := store.GetInviter(user)
	if existedInviter != "" {
		return util.BuildTextReply("抱歉，邀请码只能使用一次。")
	}

	_ = logic.AddPaidBalance(inviter, inviterReward)
	_ = store.SetInviter(user, inviter)
	return util.BuildTextReply(fmt.Sprintf(inviteSuccessMsg,
		inviterReward,
		util.GetInvitationTutorialLink(),
	))
}
