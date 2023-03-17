package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"openai/internal/config"
	"openai/internal/service/openai"
	"openai/internal/service/wechat"
	"strconv"
	"sync"
	"time"
)

var (
	success      = []byte("success")
	msgIdToReply sync.Map // K - 消息ID ， V - string
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
		answer := replyToText(msg.MsgId, msg.Content)
		echo(w, msg.GenerateEchoData(answer))
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
	answer := replyToText(intMsgId, question)
	echoJson(w, answer, "")
}

func replyToText(msgId int64, question string) string {
	v, ok := msgIdToReply.Load(msgId)
	if ok {
		return v.(string)
	}

	answerChan := make(chan string)
	leaveChan := make(chan bool)
	go func() {
		// 15s不回复微信，则失效
		answer, err := openai.Completions(question, time.Second*180)
		if err != nil {
			answer = "Try again"
		}
		msgIdToReply.Store(msgId, answer)
		select {
		case answerChan <- answer:
		case <-leaveChan:
		}
	}()

	answer := ""
	select {
	case reply := <-answerChan:
		answer = reply
		if len(answer) > 2000 {
			answer = buildAnswerURL(msgId)
		} else {
			msgIdToReply.Delete(msgId)
		}
	// 超时不要回答，会重试的
	case <-time.After(time.Second * 4):
		answer = buildAnswerURL(msgId)
		go func() {
			leaveChan <- true
		}()
	}
	return answer
}

func buildAnswerURL(msgId int64) string {
	return config.C.Wechat.MessageUrlPrefix + "/index?msgId=" + strconv.FormatInt(msgId, 10)
}

func Test(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	s := openai.Query(msg, time.Second*180)
	echoJson(w, s, "")
}

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func GetReply(w http.ResponseWriter, r *http.Request) {
	msgId := r.URL.Query().Get("msgId")
	intMsgId, err := strconv.ParseInt(msgId, 10, 64)
	if err != nil {
		panic(err)
	}
	value, ok := msgIdToReply.Load(intMsgId)
	if ok {
		echoJson(w, value.(string), "")
	} else {
		echoJson(w, "", "Reply is coming")
	}
}

func echoJson(w http.ResponseWriter, replyMsg string, errMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	var code int
	var message = replyMsg
	if errMsg != "" {
		code = -1
		message = errMsg
	}
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
