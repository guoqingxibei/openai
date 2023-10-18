package handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
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

func hitKeyword(inMsg *wechat.Msg, writer http.ResponseWriter) bool {
	question := inMsg.Content
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
			showImage(keyword, inMsg, writer)
		case help:
			showUsage(inMsg, writer)
		case transfer:
			doTransfer(inMsg, writer)
		case report:
			showReport(inMsg, writer)
		case generateCode:
			doGenerateCode(question, inMsg, writer)
		case code:
			useCodeWithPrefix(question, inMsg, writer)
		case clear:
			clearHistory(inMsg, writer)
		case invite:
			getInvitationCode(inMsg, writer)
		case reset:
			resetBalance(inMsg, writer)
		case constant.GPT3:
			fallthrough
		case constant.GPT4:
			switchGPTMode(keyword, inMsg, writer)
		}
		return true
	}

	// may hit code
	if keyword == "" {
		size := len(question)
		if size == sizeOfCode {
			invitor, _ := gptredis.GetUserByInvitationCode(strings.ToUpper(question))
			if invitor != "" {
				doInvite(invitor, inMsg, writer)
				return true
			}
		} else if size == 36 {
			codeDetailStr, _ := gptredis.FetchCodeDetail(question)
			if codeDetailStr != "" {
				useCode(codeDetailStr, inMsg, writer)
				return true
			}
		}
	}

	// missed
	return false
}

func resetBalance(inMsg *wechat.Msg, writer http.ResponseWriter) {
	userName := inMsg.FromUserName
	_ = gptredis.SetPaidBalance(userName, 0)
	_ = logic.SetBalanceOfToday(userName, 0)
	echoWechatTextMsg(writer, inMsg, "你的剩余次数已被重置。")
}

func clearHistory(inMsg *wechat.Msg, writer http.ResponseWriter) {
	_ = gptredis.DelMessages(inMsg.FromUserName)
	echoWechatTextMsg(writer, inMsg, "上下文已被清除。现在，你可以开始全新的对话啦。")
}

func useCodeWithPrefix(question string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	code := strings.Replace(question, code, "", 1)
	codeDetailStr, err := gptredis.FetchCodeDetail(code)
	if err == redis.Nil {
		echoWechatTextMsg(writer, inMsg, "无效的code。")
		return
	}

	useCode(codeDetailStr, inMsg, writer)
}

func useCode(codeDetailStr string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	var codeDetail CodeDetail
	_ = json.Unmarshal([]byte(codeDetailStr), &codeDetail)
	if codeDetail.Status == used {
		echoWechatTextMsg(writer, inMsg, "此code之前已被激活，无需重复激活。")
		return
	}

	userName := inMsg.FromUserName
	newBalance := logic.AddPaidBalance(userName, codeDetail.Times)
	codeDetail.Status = used
	codeDetailBytes, _ := json.Marshal(codeDetail)
	_ = gptredis.SetCodeDetail(codeDetail.Code, string(codeDetailBytes), false)
	echoWechatTextMsg(writer, inMsg, fmt.Sprintf("此code已被激活，额度为%d，你当前剩余的总付费次数为%d次。"+
		"回复help，可随时查看剩余次数。", codeDetail.Times, newBalance))
}

func doGenerateCode(question string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	fields := strings.Fields(question)
	if len(fields) <= 1 {
		echoWechatTextMsg(writer, inMsg, "Invalid generate-code usage")
		return
	}

	timesStr := fields[1]
	times, err := strconv.Atoi(timesStr)
	if err != nil {
		log.Printf("timesStr is %s, strconv.Atoi error is %v", timesStr, err)
		echoWechatTextMsg(writer, inMsg, "Invalid generate-code usage")
		return
	}

	quantity := 1
	if len(fields) > 2 {
		quantityStr := fields[2]
		quantity, err = strconv.Atoi(quantityStr)
		if err != nil {
			log.Printf("quantityStr is %s, strconv.Atoi error is %v", quantityStr, err)
			echoWechatTextMsg(writer, inMsg, "Invalid generate-code usage")
			return
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
		_ = gptredis.SetCodeDetail(code, string(codeDetailBytes), false)
		codes = append(codes, code)
	}
	echoWechatTextMsg(writer, inMsg, strings.Join(codes, "\n"))
}

func showReport(inMsg *wechat.Msg, writer http.ResponseWriter) {
	echoWechatTextMsg(writer, inMsg, constant.ReportInfo)
}

func showImage(keyword string, inMsg *wechat.Msg, writer http.ResponseWriter) {
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
		log.Println("wechat.GetMediaId failed", err)
		echoWechatTextMsg(writer, inMsg, constant.TryAgain)
		return
	}
	echoWechatImageMsg(writer, inMsg, QrMediaId)
}

func showUsage(inMsg *wechat.Msg, writer http.ResponseWriter) {
	userName := inMsg.FromUserName
	gptMode, _ := gptredis.GetGPTMode(userName)
	usage := fmt.Sprintf("【模式】当前模式是%s，", gptMode)
	if gptMode == constant.GPT3 {
		usage += "每次提问消耗次数1。"
	} else {
		usage += "每次提问消耗次数10。"
	}
	usage += "\n"

	usage += logic.BuildChatUsage(userName)
	balance, _ := gptredis.FetchPaidBalance(userName)
	usage += fmt.Sprintf("付费次数剩余%d次，<a href=\"%s\">点我购买次数</a>。", balance, util.GetPayLink(userName))
	usage += "\n\n<a href=\"https://cxyds.top/2023/07/03/faq.html\">更多用法</a>" +
		" | <a href=\"https://cxyds.top/2023/09/17/group-qr.html\">交流群</a>" +
		" | <a href=\"https://cxyds.top/2023/09/17/writer-qr.html\">联系作者</a>"
	echoWechatTextMsg(writer, inMsg, usage)
}

func doTransfer(inMsg *wechat.Msg, writer http.ResponseWriter) {
	if !util.AccountIsUncle() {
		echoWechatTextMsg(writer, inMsg, "此公众号不支持迁移，请移步公众号「程序员uncle」进行迁移。")
		return
	}

	userName := inMsg.FromUserName
	paidBalance, _ := gptredis.FetchPaidBalance(userName)
	reply := "你的付费次数剩余0次，无需迁移。"
	if paidBalance > 0 {
		_ = gptredis.SetPaidBalance(userName, 0)
		code := uuid.New().String()
		codeDetail := CodeDetail{
			Code:   code,
			Times:  paidBalance,
			Status: created,
		}
		codeDetailBytes, _ := json.Marshal(codeDetail)
		_ = gptredis.SetCodeDetail(code, string(codeDetailBytes), true)
		reply = fmt.Sprintf("你的付费次数剩余%d次，已在此公众号下清零。请复制下面的code发送给新公众号「程序员brother」，"+
			"即可完成迁移。感谢你的一路陪伴❤️\n\n%s", paidBalance, code)
	}
	echoWechatTextMsg(writer, inMsg, reply)
}
