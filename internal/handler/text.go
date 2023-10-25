package handler

import (
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/errorx"
	"openai/internal/service/wechat"
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

func onReceiveText(msg *message.MixMessage) (reply *message.Reply, err error) {
	if msg.MsgType == message.MsgTypeVoice {
		if msg.Recognition == "" {
			reply = util.BuildTextReply("抱歉，未识别到有效内容。")
			return
		}
		msg.Content = msg.Recognition
	}

	if len(msg.Content) > maxLengthOfQuestion {
		reply = util.BuildTextReply("哎呀，输入太长了~")
		return
	}

	hit, reply := hitKeyword(msg)
	if hit {
		return
	}

	// when WeChat server retries
	msgID := msg.MsgID
	times, _ := store.IncAccessTimes(msgID)
	if times > 1 {
		mode, _ := store.GetMode(string(msg.FromUserName))
		reply = util.BuildTextReply(buildLateDrawReply(msgID, mode))
		return
	}

	replyStr, err := genReply4Text(msg)
	if err == nil {
		reply = util.BuildTextReply(replyStr)
	}
	return
}

func genReply4Text(msg *message.MixMessage) (reply string, err error) {
	msgId := msg.MsgID
	userName := string(msg.FromUserName)
	question := strings.TrimSpace(msg.Content)
	mode, _ := store.GetMode(userName)
	ok, balanceTip := logic.DecreaseBalance(userName, mode)
	if !ok {
		reply = balanceTip
		return
	}

	drawReplyIsLate := false
	replyChan := make(chan string, 1)
	go func() {
		isVoice := msg.MsgType == message.MsgTypeVoice
		if mode == constant.Draw {
			if isVoice {
				logic.AddPaidBalance(userName, logic.GetTimesPerQuestion(mode))
				replyChan <- "draw模式下，暂不支持可能会造成误解的语音输入，请输入更准确的文字描述。"
				return
			}

			if !util.IsEnglishSentence(question) {
				logic.AddPaidBalance(userName, logic.GetTimesPerQuestion(mode))
				replyChan <- "目前，由于MidJourney对非英文的支持非常弱鸡，所以为了获得更好的用户体验，请用英文进行输入。"
				return
			}

			drawReply := logic.SubmitDrawTask(question, userName, mode)
			replyChan <- drawReply
			if drawReplyIsLate {
				err := wechat.GetAccount().
					GetCustomerMessageManager().Send(message.NewCustomerTextMessage(userName, drawReply))
				if err != nil {
					errorx.RecordError("GetCustomerMessageManager().Send() failed", err)
				}
			}
			return
		}

		err := logic.ChatCompletionStream(constant.Ohmygpt, userName, msgId, question, isVoice, mode)
		if err != nil {
			log.Printf("First ChatCompletionStream() failed, msgId is %d, error is %s", msgId, err)
			// retry
			_ = store.DelReplyChunks(msgId)
			err = logic.ChatCompletionStream(constant.OpenaiSb, userName, msgId, question, isVoice, mode)
			if err != nil {
				errorx.RecordError("Second ChatCompletionStream() failed", err)
				logic.AddPaidBalance(userName, logic.GetTimesPerQuestion(mode))
				replyChan <- constant.TryAgain
				return
			}
		}
		replyChan <- buildReply(msgId)
	}()
	select {
	case reply = <-replyChan:
	case <-time.After(time.Millisecond * 2000):
		if mode == constant.Draw {
			drawReplyIsLate = true
		}
		reply = buildLateDrawReply(msgId, mode)
	}
	return
}

func buildLateDrawReply(msgId int64, mode string) (reply string) {
	if mode == constant.Draw {
		reply = "正在提交画图任务，静候佳音..."
	} else {
		reply = buildReply(msgId)
	}
	return
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
