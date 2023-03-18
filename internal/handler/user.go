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
	"strconv"
	"time"
)

var (
	success = []byte("success")
)

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
		answer, err := replyToText(msg.MsgId, msg.Content)
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
	intMsgId, err := strconv.ParseInt(msgId, 10, 64)
	if err != nil {
		panic(err)
	}

	question := query.Get("question")
	answer, err := replyToText(intMsgId, question)
	if err == nil {
		echoJson(w, 0, answer)
	} else {
		echoJson(w, -1, "出错了，请重新提问")
	}
}

func replyToText(msgId int64, question string) (string, error) {
	msgIdStr := strconv.FormatInt(msgId, 10)
	shortMsgId, err := fetchShortMsgId(msgIdStr)
	if err != nil {
		return "", err
	}

	// indicate reply is loading
	err = setReplyToRedis(shortMsgId, "")
	if err != nil {
		return "", err
	}

	reply, err := fetchReplyFromRedis(shortMsgId)
	if err == nil && reply != "" {
		return reply, nil
	}
	if err != nil {
		return "", nil
	}

	answerChan := make(chan string)
	leaveChan := make(chan bool)
	go func() {
		// 15s不回复微信，则失效
		answer, err := openai.Completions(question, time.Second*180)
		if err != nil {
			log.Println("openai.Completions failed", err)
			answer = "出错了，请重新提问"
		}
		err = setReplyToRedis(shortMsgId, answer)
		if err != nil {
			log.Println("gptredis.Set failed", err)
			answer = "出错了，请重新提问"
		}
		select {
		case answerChan <- answer:
		case <-leaveChan:
		}
	}()

	select {
	case reply = <-answerChan:
		if len(reply) > 2000 {
			reply = buildAnswerURL(shortMsgId)
		} else {
			err := delReplyFromRedis(shortMsgId)
			if err != nil {
				log.Println("gptredis.Del failed", err)
			}
		}
	// 超时不要回答，会重试的
	case <-time.After(time.Second * 4):
		reply = buildAnswerURL(shortMsgId)
		go func() {
			leaveChan <- true
		}()
	}
	return reply, nil
}

func fetchReplyFromRedis(shortMsgId string) (string, error) {
	reply, err := gptredis.Get(buildReplyKey(shortMsgId))
	if err == nil {
		return reply, nil
	}
	return "", err
}

func setReplyToRedis(shortMsgId string, reply string) error {
	return gptredis.Set(buildReplyKey(shortMsgId), reply, time.Hour*24*7)
}

func delReplyFromRedis(shortMsgId string) error {
	return gptredis.Del(buildReplyKey(shortMsgId))
}

func buildReplyKey(shortMsgId string) string {
	return "short-msg-id:" + shortMsgId + ":reply"
}

func generateShortMsgId() (string, error) {
	shortMsgId, err := gptredis.Inc("current-max-short-id")
	if err == nil {
		return strconv.FormatInt(shortMsgId, 10), nil
	}
	return "", err
}

func fetchShortMsgId(longMsgId string) (string, error) {
	key := buildShortMsgIdKey(longMsgId)
	shortMsgId, err := gptredis.Get(key)
	if err == nil {
		return shortMsgId, nil
	}
	if err == redis.Nil {
		shortMsgId, err := generateShortMsgId()
		if err == nil {
			err := gptredis.Set(key, shortMsgId, time.Hour*24*7)
			if err == nil {
				return shortMsgId, nil
			}
			return "", err
		}
		return "", err
	}
	return "", err
}

func buildShortMsgIdKey(longMsgId string) string {
	return "long-msg-id:" + longMsgId + ":short-msg-id"
}

func buildAnswerURL(msgId string) string {
	return config.C.Wechat.MessageUrlPrefix + "/index?msgId=" + msgId
}

func Test(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	s := openai.Query(msg, time.Second*180)
	echoJson(w, 0, s)
}

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func GetReply(w http.ResponseWriter, r *http.Request) {
	shortMsgId := r.URL.Query().Get("msgId")
	reply, err := fetchReplyFromRedis(shortMsgId)
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
