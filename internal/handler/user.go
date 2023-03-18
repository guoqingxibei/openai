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
	success = []byte("success")
)

type ChatRound struct {
	question string
	answer   string
}

func WechatCheck(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	signature := query.Get("signature")
	timestamp := query.Get("timestamp")
	nonce := query.Get("nonce")
	echostr := query.Get("echostr")

	// 校验
	if wechat.CheckSignature(signature, timestamp, nonce, config.C.Wechat.Token) {
		w.Write([]byte(echostr))
		return
	}

	log.Println("此接口为公众号验证，不应该被手动调用，公众号接入校验失败")
}

// ReceiveMsg https://developers.weixin.qq.com/doc/offiaccount/Message_Management/Passive_user_reply_message.html
// 微信服务器在五秒内收不到响应会断掉连接，并且重新发起请求，总共重试三次
func ReceiveMsg(w http.ResponseWriter, r *http.Request) {
	bs, _ := io.ReadAll(r.Body)
	msg := wechat.NewMsg(bs)

	if msg == nil {
		echo(w, []byte("xml格式公众号消息接口，请勿手动调用"))
		return
	}

	// 非文本不回复(返回success表示不回复)
	switch msg.MsgType {
	// 未写的类型
	default:
		log.Printf("未实现的消息类型%s\n", msg.MsgType)
		echo(w, success)
	case "event":
		switch msg.Event {
		default:
			log.Printf("未实现的事件%s\n", msg.Event)
			echo(w, success)
		case "subscribe":
			log.Println("新增关注:", msg.FromUserName)
			echo(w, msg.GenerateEchoData(config.C.Wechat.ReplyWhenSubscribe))
			return
		case "unsubscribe":
			log.Println("取消关注:", msg.FromUserName)
			echo(w, success)
			return
		}
	case "text":
		answer, err := replyToText(msg)
		if err == nil {
			echo(w, msg.GenerateEchoData(answer))
		} else {
			echo(w, msg.GenerateEchoData("出错了，请重新提问"))
		}
	}
}

func TestReplyToText(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	msgId := query.Get("msgId")
	fromUserName := query.Get("FromUserName")
	intMsgId, err := strconv.ParseInt(msgId, 10, 64)
	if err != nil {
		panic(err)
	}

	question := query.Get("question")
	answer, err := replyToText(&wechat.Msg{
		MsgId:        intMsgId,
		Content:      question,
		FromUserName: fromUserName,
	})
	if err == nil {
		echoJson(w, 0, answer)
	} else {
		echoJson(w, -1, "出错了，请重新提问")
	}
}

func replyToText(msg *wechat.Msg) (string, error) {
	longMsgId := strconv.FormatInt(msg.MsgId, 10)
	shortMsgId, err := gptredis.FetchShortMsgId(longMsgId)
	if err != nil {
		return "", err
	}

	reply, err := gptredis.FetchReply(shortMsgId)
	answerUrl := buildAnswerURL(shortMsgId)
	if err == nil {
		if reply == "" {
			return answerUrl, nil
		}
		return reply, nil
	}
	if err != redis.Nil {
		return "", nil
	}
	// indicate reply is loading
	err = gptredis.SetReply(shortMsgId, "")
	if err != nil {
		log.Println("setReplyToRedis failed", err)
	}

	answerChan := make(chan string)
	leaveChan := make(chan bool)
	go func() {
		// 15s不回复微信，则失效
		question := strings.TrimSpace(msg.Content)
		messages, err := gptredis.FetchMessages(msg.FromUserName)
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
		answer, err := openai.Completions(messages, time.Second*180)
		if err != nil {
			log.Println("openai.Completions failed", err)
			answer = "出错了，请重新提问"
		} else {
			messages = append(messages, openai.Message{
				Role:    "assistant",
				Content: answer,
			})
			err = gptredis.SetMessages(msg.FromUserName, messages)
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
	// 超时不要回答，会重试的
	case <-time.After(time.Second * 4):
		reply = answerUrl
		go func() {
			leaveChan <- true
		}()
	}
	return reply, nil
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
	return config.C.Wechat.MessageUrlPrefix + "/index?msgId=" + msgId
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

func echo(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
