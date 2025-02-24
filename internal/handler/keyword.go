package handler

import (
	"encoding/json"
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/model"
	"openai/internal/service/errorx"
	"openai/internal/service/wechat"
	"openai/internal/store"
	"openai/internal/util"
	"strings"
)

const (
	donate   = "donate"
	group    = "group"
	help     = "help"
	contact  = "contact"
	report   = "report"
	transfer = "transfer"
	clear    = "clear"
	invite   = "invite"

	reset       = "jgq-reset"
	switchEmail = "jgq-email"
)

// prefix keyword
const (
	generateCode = "jgq-gen-code"
)

var keywords = []string{
	donate, group, help, contact, report, transfer, clear, invite, reset,
	constant.GPT3, constant.GPT4, "ds", constant.Draw, constant.Translate,
}
var keywordPrefixes = []string{generateCode, switchEmail}

func hitKeyword(msg *message.MixMessage) (hit bool, reply *message.Reply) {
	question := msg.Content
	question = strings.TrimSpace(question)
	question = strings.ToLower(question)
	var keyword string
	for _, word := range keywords {
		if question == word {
			keyword = word
			break
		}
	}
	for _, word := range keywordPrefixes {
		if strings.HasPrefix(question, word) {
			keyword = word
			break
		}
	}

	// hit keyword
	if keyword != "" {
		switch keyword {
		case contact:
			fallthrough
		case donate:
			fallthrough
		case group:
			reply = showImage(keyword)
		case help:
			reply = logic.ShowUsage(msg)
		case transfer:
			reply = logic.Transfer(msg)
		case report:
			reply = showReport()
		case generateCode:
			reply = doGenerateCode(question)
		case clear:
			reply = clearHistory(msg)
		case invite:
			reply = getInvitationCode(msg)
		case reset:
			reply = resetBalance(msg)
		case switchEmail:
			reply = switchEmailNotification(question)
		case constant.GPT3:
			fallthrough
		case constant.GPT4:
			fallthrough
		case "ds":
			keyword = constant.DeepSeekR1
			fallthrough
		case constant.Draw:
			fallthrough
		case constant.Translate:
			reply = switchMode(keyword, msg)
		}
		return true, reply
	}

	size := len(question)
	if size == sizeOfCode { // invitation code
		inviter, _ := store.GetUserByInvitationCode(strings.ToUpper(question))
		if inviter != "" {
			reply = doInvite(inviter, msg)
			return true, reply
		}
	} else if size == 36 { // code to add balance
		codeDetailStr, _ := store.GetCodeDetail(question)
		if codeDetailStr != "" {
			reply = useCode(codeDetailStr, msg)
			return true, reply
		}
	}

	// missed
	return false, nil
}

func clearHistory(msg *message.MixMessage) (reply *message.Reply) {
	_ = store.DelMessages(string(msg.FromUserName))
	return util.BuildTextReply("所有历史已被清除。现在，你可以开始全新的对话啦。")
}

func useCode(codeDetailStr string, msg *message.MixMessage) (reply *message.Reply) {
	var codeDetail model.CodeDetail
	_ = json.Unmarshal([]byte(codeDetailStr), &codeDetail)
	if codeDetail.Status == constant.Used {
		return util.BuildTextReply("此code之前已被激活，无需重复激活。")
	}

	userName := string(msg.FromUserName)
	newBalance := logic.AddPaidBalance(userName, codeDetail.Times)
	codeDetail.Status = constant.Used
	codeDetailBytes, _ := json.Marshal(codeDetail)
	_ = store.SetCodeDetail(codeDetail.Code, string(codeDetailBytes), false)
	return util.BuildTextReply(fmt.Sprintf("【激活成功】此code已被激活，额度为%d，你当前剩余的总付费次数为%d次。\n\n"+
		getShowBalanceTipWhenUseCode(), codeDetail.Times, newBalance))
}

func getShowBalanceTipWhenUseCode() string {
	if util.AccountIsUncle() {
		return "温馨提示，回复help，可查看剩余次数。"
	}
	return "温馨提示，点击菜单「次数-剩余次数」，可查看剩余次数。"
}

func showReport() (reply *message.Reply) {
	return util.BuildTextReply("bug报给jia.guoqing@qq.com，尽可能描述详细噢~")
}

func showImage(keyword string) (reply *message.Reply) {
	mediaName := constant.WriterQrImage
	switch keyword {
	case contact:
		mediaName = constant.WriterQrImage
	case group:
		mediaName = constant.GroupQrImage
	case donate:
		mediaName = constant.DonateQrImage
	}
	QrMediaId, err := wechat.GetMediaId(mediaName)
	if err != nil {
		errorx.RecordError("wechat.GetMediaId() failed", err)
		return util.BuildTextReply(constant.TryAgain)
	}
	return util.BuildImageReply(QrMediaId)
}
