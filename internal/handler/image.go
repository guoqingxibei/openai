package handler

import (
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/store"
	"openai/internal/util"
)

func onReceiveImage(msg *message.MixMessage) (reply *message.Reply) {
	return util.BuildTextReply(genReplyForImage(msg))
}

func genReplyForImage(msg *message.MixMessage) (reply string) {
	user := string(msg.FromUserName)
	mode, _ := store.GetMode(user)
	if mode != constant.GPT4 {
		tip := ""
		if util.AccountIsBrother() {
			tip = "请切换到「GPT-4对话」模式发送图片。"
		} else {
			tip = "请回复「gpt4」切换到「GPT-4对话」模式发送图片。"
		}
		return tip
	}

	ok, balanceTip := logic.DecreaseBalance(user, mode, "")
	if !ok {
		return balanceTip
	}

	_ = store.AppendReceivedImageUrl(user, msg.PicURL)
	imageUrls, _ := store.GetReceivedImageUrls(user)
	return fmt.Sprintf("【系统】已接收%d张图片，请尽快输入你的问题(关于图片的)或发送更多图片。", len(imageUrls))
}
