package handler

import (
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net/http"
	"openai/internal/config"
	"openai/internal/service/gptredis"
	"openai/internal/service/openai"
	"openai/internal/service/wechat"
	"openai/internal/util"
	"strconv"
	"strings"
	"time"
)

var (
	WechatConfig = config.C.Wechat
	success      = []byte("success")
)

type ChatRound struct {
	question string
	answer   string
}

func Check(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	signature := query.Get("signature")
	timestamp := query.Get("timestamp")
	nonce := query.Get("nonce")
	echostr := query.Get("echostr")

	// 校验
	if wechat.CheckSignature(signature, timestamp, nonce, WechatConfig.Token) {
		w.Write([]byte(echostr))
		return
	}

	log.Println("此接口为公众号验证，不应该被手动调用，公众号接入校验失败")
}

// Talk https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Passive_user_reply_message.html
// 微信服务器在五秒内收不到响应会断掉连接，并且重新发起请求，总共重试三次
func Talk(writer http.ResponseWriter, request *http.Request) {
	bs, _ := io.ReadAll(request.Body)
	inMsg := wechat.NewInMsg(bs)

	if inMsg == nil {
		echoWeChat(writer, []byte("xml格式公众号消息接口，请勿手动调用"))
		return
	}

	// 非文本不回复(返回success表示不回复)
	switch inMsg.MsgType {
	case "event":
		switch inMsg.Event {
		case "subscribe":
			log.Println("新增关注:", inMsg.FromUserName)
			echoWeChat(writer, inMsg.BuildOutMsg(WechatConfig.ReplyWhenSubscribe))
		case "unsubscribe":
			log.Println("取消关注:", inMsg.FromUserName)
			echoWeChat(writer, success)
		default:
			log.Printf("未实现的事件%s\n", inMsg.Event)
			echoWeChat(writer, success)
		}
	case "text":
		replyToText(inMsg, writer)
	default:
		log.Printf("未实现的消息类型%s\n", inMsg.MsgType)
		echoWeChat(writer, success)
	}
}

func replyToText(inMsg *wechat.Msg, writer http.ResponseWriter) {
	longMsgId := strconv.FormatInt(inMsg.MsgId, 10)
	shortMsgId, err := gptredis.FetchShortMsgId(longMsgId)
	if err != nil {
		log.Println(err)
		return
	}

	answerUrl := buildAnswerURL(shortMsgId)
	times, _ := gptredis.IncAccessTimes(shortMsgId)
	// when WeChat server retries
	if times > 1 {
		replyWhenRetry(inMsg, writer, times, shortMsgId)
		return
	}

	// when WeChat server accesses at the first time
	// indicate reply is loading
	err = gptredis.SetReply(shortMsgId, "")
	if err != nil {
		log.Println("setReplyToRedis failed", err)
	}
	answerChan := make(chan string)
	leaveChan := make(chan bool)
	go func() {
		// 15s不回复微信，则失效
		question := strings.TrimSpace(inMsg.Content)
		messages, err := gptredis.FetchMessages(inMsg.FromUserName)
		if err != nil {
			log.Println("fetchMessagesFromRedis failed", err)
			return
		}
		messages = append(messages, openai.Message{
			Role:    "user",
			Content: question,
		})
		messages, err = rotateMessages(messages)
		if err != nil {
			return
		}
		answer, err := openai.Completions(messages, time.Second*300)
		if err != nil {
			log.Println("openai.Completions failed", err)
			err = gptredis.DelReply(shortMsgId)
			if err != nil {
				log.Println("gptredis.DelReply failed", err)
			}
			answer = "出错了，请重新提问"
		} else {
			messages = append(messages, openai.Message{
				Role:    "assistant",
				Content: answer,
			})
			err = gptredis.SetMessages(inMsg.FromUserName, messages)
			if err != nil {
				log.Println("setMessagesToRedis failed", err)
			}
			err = gptredis.SetReply(shortMsgId, answer)
			if err != nil {
				log.Println("gptredis.Set failed", err)
			}
		}
		select {
		case answerChan <- answer:
		case <-leaveChan:
		}
	}()

	var reply string
	select {
	case reply = <-answerChan:
		if len(reply) > 2000 {
			reply = answerUrl
		} else {
			err := gptredis.DelReply(shortMsgId)
			if err != nil {
				log.Println("gptredis.Del failed", err)
			}
		}
		echoWechatMsg(writer, inMsg, reply)
	// wait for greater than 5s so that WeChat server retries
	case <-time.After(time.Millisecond * 5001):
		go func() {
			leaveChan <- true
		}()
	}
}

func replyWhenRetry(inMsg *wechat.Msg, writer http.ResponseWriter, times int64, shortMsgId string) {
	if times == 2 {
		pollReplyFromRedis(shortMsgId, inMsg, writer, false)
		// wait for greater than 5s so that WeChat server retries
		time.Sleep(time.Millisecond * 1001)
	} else {
		pollReplyFromRedis(shortMsgId, inMsg, writer, true)
	}
}

// poll reply from redis every second until reply is not "" in 4 seconds
func pollReplyFromRedis(shortMsgId string, inMsg *wechat.Msg, writer http.ResponseWriter, ensureFinalEcho bool) {
	cnt := 0
	for cnt < 4 {
		reply, err := gptredis.FetchReply(shortMsgId)
		if err != nil {
			log.Println(err)
			continue
		}
		if reply != "" {
			if len(reply) > 2000 {
				reply = buildAnswerURL(shortMsgId)
			}
			echoWechatMsg(writer, inMsg, reply)
			return
		}
		cnt++
		time.Sleep(time.Second)
	}
	if ensureFinalEcho {
		echoWechatMsg(writer, inMsg, buildAnswerURL(shortMsgId))
	}
}

func rotateMessages(messages []openai.Message) ([]openai.Message, error) {
	str, err := util.StringifyMessages(messages)
	for len(str) > 3000 {
		messages = messages[1:]
		str, err = util.StringifyMessages(messages)
		if err != nil {
			log.Println("stringifyMessages failed", err)
			return nil, err
		}
	}
	return messages, nil
}

func buildAnswerURL(msgId string) string {
	return WechatConfig.MessageUrlPrefix + "/index?msgId=" + msgId
}

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func GetReply(w http.ResponseWriter, r *http.Request) {
	shortMsgId := r.URL.Query().Get("msgId")
	reply, err := gptredis.FetchReply(shortMsgId)
	if err == nil {
		echoJson(w, 0, reply)
	} else if err == redis.Nil {
		echoJson(w, 1, "Not found or expired")
	} else {
		log.Println("GetReply failed", err)
		echoJson(w, 2, "Internal error")
	}
}

func echoJson(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	data, _ := json.Marshal(map[string]interface{}{
		"code":    code,
		"message": message,
	})
	w.Write(data)
}

func echoWeChat(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func echoWechatMsg(writer http.ResponseWriter, inMsg *wechat.Msg, reply string) {
	outMsg := inMsg.BuildOutMsg(reply)
	echoWeChat(writer, outMsg)
}
