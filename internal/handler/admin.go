package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/model"
	"openai/internal/service/errorx"
	"openai/internal/store"
	"openai/internal/util"
	"strconv"
	"strings"
)

func switchEmailNotification(question string) *message.Reply {
	fields := strings.Fields(question)
	if len(fields) <= 1 {
		return util.BuildTextReply("Invalid email notification switch usage")
	}

	status := fields[1]
	if status == constant.On || status == constant.Off {
		_ = store.SetEmailNotificationStatus(status)
		reply := "已开启email通知。"
		if status == constant.Off {
			reply = "已关闭email通知。"
		}
		return util.BuildTextReply(reply)
	}

	if status == constant.Status {
		status, _ := store.GetEmailNotificationStatus()
		reply := "email通知为开启状态。"
		if status == constant.Off {
			reply = "email通知为关闭状态。"
		}
		return util.BuildTextReply(reply)
	}

	return util.BuildTextReply("Invalid email notification switch usage")
}

func resetBalance(msg *message.MixMessage) (reply *message.Reply) {
	userName := string(msg.FromUserName)
	_ = store.SetPaidBalance(userName, 0)
	_ = logic.SetBalanceOfToday(userName, 0)
	return util.BuildTextReply("你的剩余次数已被重置。")
}

func doGenerateCode(question string) (reply *message.Reply) {
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
		codeDetail := model.CodeDetail{
			Code:   code,
			Times:  times,
			Status: constant.Created,
		}
		codeDetailBytes, _ := json.Marshal(codeDetail)
		_ = store.SetCodeDetail(code, string(codeDetailBytes), false)
		codes = append(codes, code)
	}
	return util.BuildTextReply(strings.Join(codes, "\n"))
}
