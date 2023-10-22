package handler

import (
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/store"
	"openai/internal/util"
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

func onReceiveText(msg *message.MixMessage) (reply *message.Reply) {
	// be compatible with voice message
	if msg.Recognition != "" {
		msg.Content = msg.Recognition
	}

	if len(msg.Content) > maxLengthOfQuestion {
		return util.BuildTextReply(constant.TooLongQuestion)
	}

	hit, reply := hitKeyword(msg)
	if hit {
		return reply
	}

	msgID := msg.MsgID
	times, _ := store.IncAccessTimes(msgID)
	// when WeChat server retries
	if times > 1 {
		return replyWhenRetry(msg)
	}

	return util.BuildTextReply(genReply4Text(msg))
}

func replyWhenRetry(msg *message.MixMessage) (reply *message.Reply) {
	return util.BuildTextReply(buildReply(msg.MsgID))
}

func genReply4Text(msg *message.MixMessage) string {
	msgId := msg.MsgID
	userName := string(msg.FromUserName)
	question := strings.TrimSpace(msg.Content)
	mode, _ := store.GetMode(userName)
	ok, balanceTip := logic.CheckBalance(userName, mode)
	if !ok {
		return balanceTip
	}

	replyChan := make(chan string, 1)
	go func() {
		isVoice := msg.Recognition != ""
		err := logic.ChatCompletionStream(constant.Ohmygpt, userName, msgId, question, isVoice, mode)
		if err != nil {
			log.Printf("[%d] logic.ChatCompletionStream with Ohmygpt failed %s", msgId, err)
			// retry with api2d vendor
			_ = store.DelReplyChunks(msgId)
			err = logic.ChatCompletionStream(constant.OpenaiSb, userName, msgId, question, isVoice, mode)
			if err != nil {
				log.Printf("[%d] logic.ChatCompletionStream with OpenaiSb failed %s", msgId, err)
				logic.RecordError(err)
				replyChan <- constant.TryAgain
				return
			}
		}

		err = logic.DecrBalanceOfToday(userName, mode)
		if err != nil {
			log.Println("store.DecrBalance failed", err)
		}
		replyChan <- buildReply(msgId)
	}()
	select {
	case reply := <-replyChan:
		return reply
	case <-time.After(time.Millisecond * 2000):
		return buildReply(msgId)
	}
}

func buildReply(msgId int64) string {
	reply, reachEnd := logic.FetchReply(msgId)
	if len(reply) > maxLengthOfReply {
		reply = buildReplyWithShowMore(string([]rune(reply)[:maxRuneLengthOfReply]), msgId)
	} else {
		if reachEnd {
			// Intent to display internal images via web
			if strings.Contains(reply, "![](./images/") {
				runes := []rune(reply)
				length := len(runes)
				if length > 30 {
					length = 30
				}
				reply = buildReplyWithShowMore(string(runes[:length]), msgId)
			}
		} else {
			if reply == "" {
				reply = buildReplyURL(msgId, "查看回复")
			} else {
				reply = buildReplyWithShowMore(reply, msgId)
			}
		}
	}
	return reply
}

func buildReplyWithShowMore(answer string, msgId int64) string {
	return trimTailingPuncts(answer) + "..." + buildReplyURL(msgId, "更多")
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

func buildReplyURL(msgId int64, desc string) string {
	url := config.C.Wechat.MessageUrlPrefix + "/answer/#/?msgId=" + strconv.FormatInt(msgId, 10)
	return fmt.Sprintf("<a href=\"%s\">%s</a>", url, desc)
}
