package handler

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/gptredis"
	"openai/internal/service/openai"
	"openai/internal/service/wechat"
	"strconv"
	"strings"
	"time"
)

const (
	maxLengthOfReply = 4090
)

func echoText(inMsg *wechat.Msg, writer http.ResponseWriter) {
	// be compatible with voice message
	if inMsg.Recognition != "" {
		inMsg.Content = inMsg.Recognition
	}
	if hitKeyword(inMsg, writer) {
		return
	}

	msgId := inMsg.MsgId
	times, _ := gptredis.IncAccessTimes(msgId)
	// when WeChat server retries
	if times > 1 {
		replyWhenRetry(inMsg, writer, times)
		return
	}

	// set empty string when WeChat server accesses at the first time
	// to indicate reply is loading
	err := logic.SetEmptyReply(msgId)
	if err != nil {
		log.Println("replylogic.SetEmptyReply failed", err)
	}
	outChan := make(chan []byte, 1)
	go func() {
		out := genAnswer4Text(inMsg)
		outChan <- out
	}()

	var out []byte
	select {
	case out = <-outChan:
		echoWeChat(writer, out)
	// wait for greater than 5s so that WeChat server retries
	case <-time.After(time.Millisecond * 5001):
	}
}

func genAnswer4Text(inMsg *wechat.Msg) []byte {
	var out []byte
	msgId := inMsg.MsgId
	userName := inMsg.FromUserName
	question := strings.TrimSpace(inMsg.Content)
	answerUrl := buildAnswerURL(msgId)
	mode, err := gptredis.FetchModeForUser(userName)
	if err != nil {
		if err != redis.Nil {
			log.Println("ptredis.FetchModeForUser failed", err)
		}
		mode = constant.Chat
	}
	ok, out := logic.CheckBalance(inMsg, mode)
	if !ok {
		return out
	}
	if mode == constant.Chat {
		answer, err := logic.ChatCompletion(userName, question)
		if err != nil {
			log.Println("openai.ChatCompletions failed", err)
			err = gptredis.DelReply(msgId)
			if err != nil {
				log.Println("gptredis.DelReply failed", err)
			}
			out = inMsg.BuildTextMsg(constant.TryAgain)
		} else {
			_, err := logic.DecrBalanceOfToday(userName, mode)
			if err != nil {
				log.Println("gptredis.DecrBalance failed", err)
			}
			answer = logic.AppendIfPossible(userName, answer)
			answer = prependRecognition(inMsg, answer)
			err = logic.SetTextReply(msgId, answer)
			if err != nil {
				log.Println("replylogic.SetReply failed", err)
			}
			if len(answer) > maxLengthOfReply {
				answer = answerUrl
			}
			out = inMsg.BuildTextMsg(answer)
		}
	} else {
		url, err := openai.GenerateImage(question)
		if err != nil {
			out = inMsg.BuildTextMsg(err.Error())
			err = gptredis.DelReply(msgId)
			if err != nil {
				log.Println("gptredis.DelReply failed", err)
			}
		} else {
			_, err := logic.DecrBalanceOfToday(userName, mode)
			if err != nil {
				log.Println("gptredis.DecrBalance failed", err)
			}
			err = logic.SetImageReply(msgId, url, "")
			if err != nil {
				log.Println("replylogic.SetImageReply failed", err)
			}
			mediaId, err := wechat.UploadImageFromUrl(url)
			if err != nil {
				log.Println("wechat.UploadImageFromUrl failed", err)
				out = inMsg.BuildTextMsg(url)
			} else {
				err = logic.SetImageReply(msgId, url, mediaId)
				if err != nil {
					log.Println("replylogic.SetImageReply failed", err)
				}
				out = inMsg.BuildImageMsg(mediaId)
			}
		}
	}
	return out
}

func replyWhenRetry(inMsg *wechat.Msg, writer http.ResponseWriter, times int64) {
	if times == 2 {
		// wait for greater than 5s so that WeChat server retries
		pollReplyFromRedis(51, inMsg, writer)
	} else {
		pollReplyFromRedis(40, inMsg, writer)
	}
}

// poll reply from redis every 0.1 second until reply is not "" in 5 seconds
func pollReplyFromRedis(pollCnt int, inMsg *wechat.Msg, writer http.ResponseWriter) {
	cnt := 0
	msgId := inMsg.MsgId
	for cnt < pollCnt {
		cnt++
		reply, err := logic.FetchReply(msgId)
		if err != nil {
			log.Println("gptredis.FetchReply failed", err)
			continue
		}
		replyType := reply.ReplyType
		if replyType != "" {
			if replyType == logic.Text {
				content := reply.Content
				if len(content) > maxLengthOfReply {
					content = buildAnswerURL(msgId)
				}
				echoWechatTextMsg(writer, inMsg, content)
				return
			} else {
				mediaId := reply.MediaId
				if mediaId != "" {
					echoWechatImageMsg(writer, inMsg, mediaId)
					return
				}
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
	echoWechatTextMsg(writer, inMsg, buildAnswerURL(msgId))
}

func prependRecognition(inMsg *wechat.Msg, content string) string {
	if inMsg.Recognition != "" {
		content = "「" + inMsg.Recognition + "」\n\n" + content
	}
	return content
}

func buildAnswerURL(msgId int64) string {
	url := config.C.Wechat.MessageUrlPrefix + "/index?msgId=" + strconv.FormatInt(msgId, 10)
	return fmt.Sprintf("<a href=\"%s\">点击查看回复</a>", url)
}

func echoWechatImageMsg(writer http.ResponseWriter, inMsg *wechat.Msg, mediaId string) {
	outMsg := inMsg.BuildImageMsg(mediaId)
	echoWeChat(writer, outMsg)
}
