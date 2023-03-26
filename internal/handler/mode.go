package handler

import (
	"log"
	"net/http"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
)

// mode
const (
	Chat  = "chat"
	Image = "image"
)

var modes = [2]string{Chat, Image}

func checkModeSwitch(question string) (string, bool) {
	for _, mode := range modes {
		if question == mode {
			return mode, true
		}
	}
	return "", false
}

func setModeThenReply(mode string, inMsg *wechat.Msg, writer http.ResponseWriter) {
	err := gptredis.SetModeForUser(inMsg.FromUserName, mode)
	if err != nil {
		log.Println("gptredis.SetModeForUser failed", err)
		echoWechatTextMsg(writer, inMsg, constant.TryAgain)
	} else {
		echoWechatTextMsg(writer, inMsg, buildReplyForMode(mode))
	}
}

func buildReplyForMode(mode string) string {
	reply := "已切换到 " + mode + " 模式。"
	if mode == Image {
		reply += "\n\n用法：你说一句图片描述，我回一张图片，单轮对话，每天仅限 5 次。\n\n说明：图片生成价格昂贵，仅仅 Image API 调用就单次 2 毛，作者 hold 不住，请理解。\n\n回复 help 查看使用方法。如果觉得体验不错，可回复 donate 捐赠作者。"
	} else {
		reply += "\n\n用法：你说一句，我回一句，多轮对话，不限次数。\n\n说明：此 Chat API 价格便宜，作者能 hold 住。\n\n回复 help 查看使用方法。如果觉得体验不错，可回复 donate 捐赠作者。"
	}
	return reply
}
