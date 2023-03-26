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
	usage := "当前是 " + mode + " 模式。"
	usage += "\n\n回复 chat，开启 chat 模式。此模式是默认模式，在此模式下，你问我答，多轮对话，不限次数。"
	usage += "\n\n回复 image，开启 image 模式。在此模式下，你说一句尽可能完整的图片描述，我画一张对应的图片，单轮对话，每天仅限 5 次。"
	usage += "\n\n回复 donate，可对作者进行捐赠。所有对话都会产生费用，你的捐赠可以减轻作者的财务压力。但捐赠与否，并不影响你对任何服务的使用。"
	echoWechatTextMsg(writer, inMsg, usage)
}
