package handler

import (
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
)

func showUsage(inMsg *wechat.Msg, writer http.ResponseWriter) {
	mode, err := gptredis.FetchModeForUser(inMsg.FromUserName)
	if err != nil {
		if err != redis.Nil {
			log.Println("gptredis.FetchModeForUser failed", err)
			echoWechatTextMsg(writer, inMsg, constant.TryAgain)
			return
		}
		mode = Chat
	}
	usage := "你当前是 " + mode + " 模式。"
	usage += "\n\n回复 chat，可开启不限次数的 chat 模式，此模式为默认模式。"
	usage += "\n回复 image，可开启每天仅限 5 次的 image 模式。"
	usage += "\n回复 donate，可对作者进行捐赠。"
	echoWechatTextMsg(writer, inMsg, usage)
}
