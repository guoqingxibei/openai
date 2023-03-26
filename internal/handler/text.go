package handler

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"openai/internal/constant"
	appendlogic "openai/internal/logic/append"
	openailogic "openai/internal/logic/openai"
	replylogic "openai/internal/logic/reply"
	"openai/internal/service/gptredis"
	"openai/internal/service/openai"
	"openai/internal/service/wechat"
	"strconv"
	"strings"
	"time"
)

const (
	maxLengthOfReply = 4000
)

func echoText(inMsg *wechat.Msg, writer http.ResponseWriter) {
	question := strings.TrimSpace(inMsg.Content)
	if hitKeyword(inMsg, writer) {
		return
	}

	if mode, ok := checkModeSwitch(question); ok {
		setModeThenReply(mode, inMsg, writer)
		return
	}

	msgId := inMsg.MsgId
	times, _ := gptredis.IncAccessTimes(msgId)
	// when WeChat server retries
	if times > 1 {
		replyWhenRetry(inMsg, writer, times)
		return
	}

	// when WeChat server accesses at the first time
	// indicate reply is loading
	err := replylogic.SetEmptyReply(msgId)
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
	msgId := inMsg.MsgId
	userName := inMsg.FromUserName
	question := strings.TrimSpace(inMsg.Content)
	answerUrl := buildAnswerURL(msgId)
	mode, err := gptredis.FetchModeForUser(userName)
	if err != nil {
		if err != redis.Nil {
			log.Println("ptredis.FetchModeForUser failed", err)
		}
		mode = Chat
	}
	var out []byte
	if mode == Chat {
		answer, err := openailogic.ChatCompletion(userName, question)
		if err != nil {
			log.Println("openai.ChatCompletions failed", err)
			err = gptredis.DelReply(msgId)
			out = inMsg.BuildTextMsg(constant.TryAgain)
		} else {
			answer = appendlogic.AppendHelpDescIfPossible(userName, answer)
			err = replylogic.SetTextReply(msgId, answer)
			if err != nil {
				log.Println("replylogic.SetReply failed", err)
			}
		}
		if len(answer) > maxLengthOfReply {
			answer = answerUrl
		}
		out = inMsg.BuildTextMsg(answer)
	} else {
		balance := openailogic.FetchImageBalance(userName)
		if balance <= 0 {
			return inMsg.BuildTextMsg("你今天的图片次数已用完，24 小时后再来吧。\n\n回复 chat，可切换到不限次数的 chat 模式。")
		}
		url, err := openai.GenerateImage(question)
		if err != nil {
			log.Println("openai.GenerateImage failed", err)
			err = gptredis.DelReply(msgId)
			out = inMsg.BuildTextMsg(constant.TryAgain)
		} else {
			_, err := gptredis.DecrImageBalance(userName)
			if err != nil {
				log.Println("gptredis.DecrImageBalance failed", err)
			}
			err = replylogic.SetImageReply(msgId, url, "")
			if err != nil {
				log.Println("replylogic.SetImageReply failed", err)
			}
			mediaId, err := wechat.UploadImageFromUrl(url)
			if err != nil {
				log.Println("wechat.UploadImageFromUrl failed", err)
				out = inMsg.BuildTextMsg(url)
			} else {
				err = replylogic.SetImageReply(msgId, url, mediaId)
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
		reply, err := replylogic.FetchReply(msgId)
		if err != nil {
			log.Println("gptredis.FetchReply failed", err)
			continue
		}
		replyType := reply.ReplyType
		if replyType != "" {
			if replyType == replylogic.Text {
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

func buildAnswerURL(msgId int64) string {
	url := wechatConfig.MessageUrlPrefix + "/index?msgId=" + strconv.FormatInt(msgId, 10)
	return fmt.Sprintf("<a href=\"%s\">点击查看回复</a>", url)
}

func echoWechatImageMsg(writer http.ResponseWriter, inMsg *wechat.Msg, mediaId string) {
	outMsg := inMsg.BuildImageMsg(mediaId)
	echoWeChat(writer, outMsg)
}
