package handler

import (
	"fmt"
	"net/http"
	"openai/internal/logic"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"time"
)

const inviteTutorial = `【邀请码】
%s

【邀请流程】
1. 分享公众号给你的朋友关注
2. 让ta向公众号发送此邀请码

注意，邀请码可以被多人使用，但只能在关注后30分钟内使用。

【邀请奖励】
每次邀请成功，将为你充值20次的额度，为ta充值10次的额度。

额度永久有效，视作付费额度，可在菜单栏点击「剩余次数」查看。
`
const inviteSuccessMsg = "【成功接受邀请】已为你的邀请者充值20次的额度，为你充值10次的额度，你当前的付费额度剩余%d次。"

var codeChars = []rune{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K',
				'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
var base = len(codeChars) // char count
const sizeOfCode = 6      // the length of invitation code
const halfAnHour = 30 * 60

func getInvitationCode(inMsg *wechat.Msg, writer http.ResponseWriter) {
	user := inMsg.FromUserName
	code, _ := gptredis.GetInvitationCode(user)
	if code == "" {
		cursor, _ := gptredis.IncInvitationCodeCursor()
		code = convertToInvitationCode(int(cursor - 1))
		_ = gptredis.SetInvitationCode(user, code)
		_ = gptredis.SetUserByInvitationCode(code, user)
	}
	echoWechatTextMsg(writer, inMsg, fmt.Sprintf(inviteTutorial, code))
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

func doInvite(invitor string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	user := inMsg.FromUserName
	if user == invitor {
		echoWechatTextMsg(writer, inMsg, "抱歉，你无法使用自己的邀请码。")
		return
	}

	currentTimestamp := time.Now().Unix()
	subScribeTimestamp, _ := gptredis.FetchSubscribeTimestamp(user)
	if currentTimestamp-subScribeTimestamp > halfAnHour {
		echoWechatTextMsg(writer, inMsg, "抱歉，邀请码仅在关注公众号半小时内输入有效。")
		return
	}

	existedInvitor, _ := gptredis.GetInvitor(user)
	if existedInvitor != "" {
		echoWechatTextMsg(writer, inMsg, "抱歉，邀请码只能使用一次。")
		return
	}

	_ = logic.AddPaidBalance(invitor, 20)
	userPaidBalance := logic.AddPaidBalance(user, 10)
	_ = gptredis.SetInvitor(user, invitor)
	echoWechatTextMsg(writer, inMsg, fmt.Sprintf(inviteSuccessMsg, userPaidBalance))
}
