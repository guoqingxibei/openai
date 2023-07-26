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
)

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

var keywords = []string{donate, group, help, contact, report, transfer}
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
		}
		return true
	}

	// may hit code
	if keyword == "" && len(question) == 36 {
		codeDetailStr, _ := gptredis.FetchCodeDetail(question)
		if codeDetailStr != "" {
			useCode(codeDetailStr, inMsg, writer)
			return true
		}
	}

	// missed
	return false
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
	balance, _ := gptredis.FetchPaidBalance(userName)
	_ = gptredis.SetPaidBalance(userName, codeDetail.Times+balance)
	codeDetail.Status = used
	codeDetailBytes, _ := json.Marshal(codeDetail)
	_ = gptredis.SetCodeDetail(codeDetail.Code, string(codeDetailBytes), false)
	echoWechatTextMsg(writer, inMsg, fmt.Sprintf("此code已被激活，额度为%d。回复help，可随时查看剩余次数。", codeDetail.Times))
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
	usage := logic.BuildChatUsage(userName)
	balance, _ := gptredis.FetchPaidBalance(userName)
	usage += fmt.Sprintf("付费次数剩余%d次，<a href=\"brother.cxyds.top/shop?uncle_openid=%s\">点我可购买次数</a>。", balance, userName)
	usage += "\n\n" + constant.HelpDesc
	echoWechatTextMsg(writer, inMsg, usage)
}

func doTransfer(inMsg *wechat.Msg, writer http.ResponseWriter) {
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
