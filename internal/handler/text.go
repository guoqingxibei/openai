package handler

import (
	"errors"
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/errorx"
	"openai/internal/service/wechat"
	"openai/internal/store"
	"openai/internal/util"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	maxLengthOfReply     = 4000
	maxRuneLengthOfReply = 200
	maxLengthOfQuestion  = 3000 // ~ 1000 Chinese characters
)

func onReceiveText(msg *message.MixMessage) (reply *message.Reply) {
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
	times, _ := store.IncRequestTimesForMsg(msgID)
	if times > 1 {
		mode, _ := store.GetMode(string(msg.FromUserName))
		reply = util.BuildTextReply(buildLateReply(msgID, mode))
		return
	}

	reply = util.BuildTextReply(genReply4Text(msg))
	return
}

func genReply4Text(msg *message.MixMessage) (reply string) {
	msgId := msg.MsgID
	user := string(msg.FromUserName)
	question := strings.TrimSpace(msg.Content)
	mode, _ := store.GetMode(user)

	isVoice := msg.MsgType == message.MsgTypeVoice
	if isVoice && (mode == constant.Draw || mode == constant.TTS) {
		return fmt.Sprintf("「%s」模式仅支持文字输入。", logic.GetModeName(mode))
	}

	ok, balanceTip := logic.DecreaseBalance(user, mode, question)
	if !ok {
		reply = balanceTip
		return
	}

	replyIsLate := false
	replyChan := make(chan string, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicMsg := fmt.Sprintf("%v\n%s", r, debug.Stack())
				errorx.RecordError("failed due to a panic", errors.New(panicMsg))
			}
		}()

		if mode == constant.Draw {
			drawReply := logic.SubmitDrawTask(question, user, mode)
			replyChan <- drawReply
			if replyIsLate {
				err := wechat.GetAccount().GetCustomerMessageManager().
					Send(message.NewCustomerTextMessage(user, drawReply))
				if err != nil {
					errorx.RecordError("GetCustomerMessageManager().Send() failed", err)
				}
			}
			return
		}

		if mode == constant.GPT3 || mode == constant.GPT4 {
			// convert voice to text
			if isVoice {
				textResult, err := logic.GetTextFromVoice(msg.MediaID)
				if err != nil {
					errorx.RecordError("GetTextFromVoice() failed", err)
				}
				if textResult == "" {
					replyChan <- "抱歉，未识别到有效内容。"
					return
				}
				question = textResult
			}

			logic.CreateChatStreamEx(user, msgId, question, isVoice, mode)
			replyChan <- buildReplyForChat(msgId)
			return
		}

		if mode == constant.TTS {
			ttsReply := logic.TextToVoiceEx(question, user)
			replyChan <- ttsReply
			if replyIsLate && ttsReply != "" {
				err := wechat.GetAccount().GetCustomerMessageManager().
					Send(message.NewCustomerTextMessage(user, ttsReply))
				if err != nil {
					errorx.RecordError("GetCustomerMessageManager().Send() failed", err)
				}
			}
			return
		}

		// Unknown mode
		replyChan <- constant.TryAgain
	}()
	select {
	case reply = <-replyChan:
	case <-time.After(time.Millisecond * 3000):
		if mode == constant.Draw || mode == constant.TTS {
			replyIsLate = true
		}
		reply = buildLateReply(msgId, mode)
	}
	return
}

func buildLateReply(msgId int64, mode string) (reply string) {
	if mode == constant.Draw {
		reply = "正在提交绘画任务，静候佳音..."
	} else if mode == constant.TTS {
		reply = "转换中，静候佳音..."
	} else {
		reply = buildReplyForChat(msgId)
	}
	return
}

func buildReplyForChat(msgId int64) string {
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
	return strings.TrimSpace(reply)
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
