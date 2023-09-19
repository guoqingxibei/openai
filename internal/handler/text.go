package handler

import (
	"fmt"
	"log"
	"net/http"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	maxLengthOfReply     = 4000
	maxRuneLengthOfReply = 200
	maxLengthOfQuestion  = 2000
)

func echoText(inMsg *wechat.Msg, writer http.ResponseWriter) {
	// be compatible with voice message
	if inMsg.Recognition != "" {
		inMsg.Content = inMsg.Recognition
	}

	if len(inMsg.Content) > maxLengthOfQuestion {
		echoWechatTextMsg(writer, inMsg, constant.TooLongQuestion)
		return
	}

	if hitKeyword(inMsg, writer) {
		return
	}

	msgId := inMsg.MsgId
	times, _ := gptredis.IncAccessTimes(msgId)
	// when WeChat server retries
	if times > 1 {
		replyWhenRetry(inMsg, writer)
		return
	}

	echoWechatTextMsg(writer, inMsg, genAnswer4Text(inMsg))
}

func replyWhenRetry(inMsg *wechat.Msg, writer http.ResponseWriter) {
	echoWechatTextMsg(writer, inMsg, buildAnswer(inMsg.MsgId))
}

func genAnswer4Text(inMsg *wechat.Msg) string {
	msgId := inMsg.MsgId
	userName := inMsg.FromUserName
	question := strings.TrimSpace(inMsg.Content)
	gptMode, _ := gptredis.GetGPTMode(userName)
	ok, reply := logic.CheckBalance(inMsg, gptMode)
	if !ok {
		return reply
	}

	answerChan := make(chan string, 1)
	go func() {
		isVoice := inMsg.Recognition != ""
		err := logic.ChatCompletionStream(constant.OpenaiSb, userName, msgId, question, isVoice, gptMode)
		if err != nil {
			log.Println("logic.ChatCompletionStream with OpenaiSb failed", err)
			// retry with api2d vendor
			_ = gptredis.DelReplyChunks(msgId)
			err = logic.ChatCompletionStream(constant.OpenaiApi2d, userName, msgId, question, isVoice, gptMode)
			if err != nil {
				log.Println("logic.ChatCompletionStream with OpenaiApi2d failed", err)
				logic.RecordError(err)
				answerChan <- constant.TryAgain
				return
			}
		}

		err = logic.DecrBalanceOfToday(userName, gptMode)
		if err != nil {
			log.Println("gptredis.DecrBalance failed", err)
		}
		answerChan <- buildAnswer(msgId)
	}()
	select {
	case answer := <-answerChan:
		return answer
	case <-time.After(time.Millisecond * 2000):
		return buildAnswer(msgId)
	}
}

func buildAnswer(msgId int64) string {
	answer, reachEnd := logic.FetchAnswer(msgId)
	if len(answer) > maxLengthOfReply {
		answer = buildAnswerWithShowMore(string([]rune(answer)[:maxRuneLengthOfReply]), msgId)
	} else {
		if reachEnd {
			// Intent to display internal images via web
			if strings.Contains(answer, "![](./images/") {
				runes := []rune(answer)
				length := len(runes)
				if length > 30 {
					length = 30
				}
				answer = buildAnswerWithShowMore(string(runes[:length]), msgId)
			} else {
				answer = answer + "\n" + buildAnswerURL(msgId, "查看网页版")
			}
		} else {
			if answer == "" {
				answer = buildAnswerURL(msgId, "点击查看回复")
			} else {
				answer = buildAnswerWithShowMore(answer, msgId)
			}
		}
	}
	return answer
}

func buildAnswerWithShowMore(answer string, msgId int64) string {
	return trimTailingPuncts(answer) + "...\n" + buildAnswerURL(msgId, "查看更多")
}

func trimTailingPuncts(answer string) string {
	runeAnswer := []rune(answer)
	if len(runeAnswer) <= 0 {
		return ""
	}
	tailIdx := -1
	for i := len(runeAnswer) - 1; i >= 0; i-- {
		if !unicode.IsPunct(runeAnswer[i]) {
			tailIdx = i
			break
		}
	}
	return string(runeAnswer[:tailIdx+1])
}

func buildAnswerURL(msgId int64, desc string) string {
	url := config.C.Wechat.MessageUrlPrefix + "/answer/#/?msgId=" + strconv.FormatInt(msgId, 10)
	return fmt.Sprintf("<a href=\"%s\">%s</a>", url, desc)
}

func echoWechatImageMsg(writer http.ResponseWriter, inMsg *wechat.Msg, mediaId string) {
	outMsg := inMsg.BuildImageMsg(mediaId)
	echoWeChat(writer, outMsg)
}
