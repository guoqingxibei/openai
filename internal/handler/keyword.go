package handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/errorx"
	"openai/internal/service/wechat"
	"openai/internal/store"
	"openai/internal/util"
	"strconv"
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
	reset    = "jgq-reset"
)

// prefix keyword
const (
	generateCode = "generate-code"
	code         = "code:"
)

type CodeDetail struct {
	Code   string `json:"code"`
	Times  int    `json:"times"`
	Status string `json:"status"`
}

const (
	created = "created"
	used    = "used"
)

var keywords = []string{
	donate, group, help, contact, report, transfer, clear, invite, reset, constant.GPT3, constant.GPT4,
}
var keywordPrefixes = []string{generateCode, code}

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
			reply = showUsage(msg)
		case transfer:
			reply = doTransfer(msg)
		case report:
			reply = showReport()
		case generateCode:
			reply = doGenerateCode(question, msg)
		case code:
			reply = useCodeWithPrefix(question, msg)
		case clear:
			reply = clearHistory(msg)
		case invite:
			reply = getInvitationCode(msg)
		case reset:
			reply = resetBalance(msg)
		case constant.GPT3:
			fallthrough
		case constant.GPT4:
			reply = switchMode(keyword, msg)
		}
		return true, reply
	}

	// may hit code
	if keyword == "" {
		size := len(question)
		if size == sizeOfCode {
			inviter, _ := store.GetUserByInvitationCode(strings.ToUpper(question))
			if inviter != "" {
				reply = doInvite(inviter, msg)
				return true, reply
			}
		} else if size == 36 {
			codeDetailStr, _ := store.GetCodeDetail(question)
			if codeDetailStr != "" {
				reply = useCode(codeDetailStr, msg)
				return true, reply
			}
		}
	}

	// missed
	return false, nil
}

func resetBalance(msg *message.MixMessage) (reply *message.Reply) {
	userName := string(msg.FromUserName)
	_ = store.SetPaidBalance(userName, 0)
	_ = logic.SetBalanceOfToday(userName, 0)
	return util.BuildTextReply("你的剩余次数已被重置。")
}

func clearHistory(msg *message.MixMessage) (reply *message.Reply) {
	_ = store.DelMessages(string(msg.FromUserName))
	return util.BuildTextReply("上下文已被清除。现在，你可以开始全新的对话啦。")
}

func useCodeWithPrefix(question string, msg *message.MixMessage) (reply *message.Reply) {
	code := strings.Replace(question, code, "", 1)
	codeDetailStr, err := store.GetCodeDetail(code)
	if err == redis.Nil {
		return util.BuildTextReply("无效的code。")
	}

	return useCode(codeDetailStr, msg)
}

func useCode(codeDetailStr string, msg *message.MixMessage) (reply *message.Reply) {
	var codeDetail CodeDetail
	_ = json.Unmarshal([]byte(codeDetailStr), &codeDetail)
	if codeDetail.Status == used {
		return util.BuildTextReply("此code之前已被激活，无需重复激活。")
	}

	userName := string(msg.FromUserName)
	newBalance := logic.AddPaidBalance(userName, codeDetail.Times)
	codeDetail.Status = used
	codeDetailBytes, _ := json.Marshal(codeDetail)
	_ = store.SetCodeDetail(codeDetail.Code, string(codeDetailBytes), false)
	return util.BuildTextReply(fmt.Sprintf("【激活成功】此code已被激活，额度为%d，你当前剩余的总付费次数为%d次。"+
		getShowBalanceTipWhenUseCode(), codeDetail.Times, newBalance))
}

func getShowBalanceTipWhenUseCode() string {
	if util.AccountIsUncle() {
		return "回复help，可查看剩余次数。"
	}
	return "点击菜单「次数-剩余次数」，可查看剩余次数。"
}

func doGenerateCode(question string, msg *message.MixMessage) (reply *message.Reply) {
	fields := strings.Fields(question)
	if len(fields) <= 1 {
		return util.BuildTextReply("Invalid generate-code usage")
	}

	timesStr := fields[1]
	times, err := strconv.Atoi(timesStr)
	if err != nil {
		errorx.RecordError("strconv.Atoi() failed", err)
		return util.BuildTextReply("Invalid generate-code usage")
	}

	quantity := 1
	if len(fields) > 2 {
		quantityStr := fields[2]
		quantity, err = strconv.Atoi(quantityStr)
		if err != nil {
			log.Printf("quantityStr is %s, strconv.Atoi error is %v", quantityStr, err)
			return util.BuildTextReply("Invalid generate-code usage")
		}
	}

	var codes []string
	for i := 0; i < quantity; i++ {
		code := uuid.New().String()
		codeDetail := CodeDetail{
			Code:   code,
			Times:  times,
			Status: created,
		}
		codeDetailBytes, _ := json.Marshal(codeDetail)
		_ = store.SetCodeDetail(code, string(codeDetailBytes), false)
		codes = append(codes, code)
	}
	return util.BuildTextReply(strings.Join(codes, "\n"))
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

func showUsage(msg *message.MixMessage) (reply *message.Reply) {
	userName := string(msg.FromUserName)
	mode, _ := store.GetMode(userName)
	usage := fmt.Sprintf("【模式】当前模式是%s，每次对话消耗次数%d。\n", mode, logic.GetTimesPerQuestion(mode))

	usage += logic.BuildChatUsage(userName)
	balance, _ := store.GetPaidBalance(userName)
	usage += fmt.Sprintf("付费次数剩余%d次，可以<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。",
		balance,
		util.GetPayLink(userName),
		util.GetInvitationTutorialLink(),
	)
	usage += "\n\n<a href=\"https://cxyds.top/2023/07/03/faq.html\">更多用法</a>" +
		" | <a href=\"https://cxyds.top/2023/09/17/group-qr.html\">交流群</a>" +
		" | <a href=\"https://cxyds.top/2023/09/17/writer-qr.html\">联系作者</a>"
	return util.BuildTextReply(usage)
}

func doTransfer(msg *message.MixMessage) (reply *message.Reply) {
	if !util.AccountIsUncle() {
		return util.BuildTextReply("此公众号不支持迁移，请移步公众号「程序员uncle」进行迁移。")
	}

	userName := string(msg.FromUserName)
	paidBalance, _ := store.GetPaidBalance(userName)
	replyText := "你的付费次数剩余0次，无需迁移。"
	if paidBalance > 0 {
		_ = store.SetPaidBalance(userName, 0)
		code := uuid.New().String()
		codeDetail := CodeDetail{
			Code:   code,
			Times:  paidBalance,
			Status: created,
		}
		codeDetailBytes, _ := json.Marshal(codeDetail)
		_ = store.SetCodeDetail(code, string(codeDetailBytes), true)
		replyText = fmt.Sprintf("你的付费次数剩余%d次，已在此公众号下清零。请复制下面的code发送给新公众号「程序员brother」，"+
			"即可完成迁移。感谢你的一路陪伴❤️\n\n%s", paidBalance, code)
	}
	return util.BuildTextReply(replyText)
}
