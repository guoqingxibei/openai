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
	ok, msg := logic.CheckBalance(inMsg, constant.Chat)
	if !ok {
		return msg
	}

	answerChan := make(chan string, 1)
	go func() {
		err := logic.ChatCompletionStream(userName, msgId, question, inMsg.Recognition != "")
		if err != nil {
			log.Println("logic.ChatCompletionStream error", err)
			answerChan <- constant.TryAgain
		} else {
			err := logic.DecrBalanceOfToday(userName, constant.Chat)
			if err != nil {
				log.Println("gptredis.DecrBalance failed", err)
			}
			answerChan <- buildAnswer(msgId)
		}
	}()
	select {
	case answer := <-answerChan:
		return answer
	case <-time.After(time.Millisecond * 4500):
		return buildAnswer(msgId)
	}
}

func buildAnswer(msgId int64) string {
	answer, reachEnd := logic.FetchAnswer(msgId)
	if len(answer) > maxLengthOfReply {
		answer = buildAnswerWithShowMore(trimAnswerAsRune(answer), msgId)
	} else {
		if reachEnd {
			answer = answer + "\n" + buildAnswerURL(msgId, "查看网页版")
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

func trimAnswerAsRune(answer string) string {
	return string([]rune(answer)[:maxRuneLengthOfReply])
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
