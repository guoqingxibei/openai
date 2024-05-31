package handler

import (
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/errorx"
	"openai/internal/store"
	"openai/internal/util"
)

const (
	maxReceivedImages = 4
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

	imageUrls, _ := store.GetReceivedImageUrls(user)
	count := len(imageUrls)
	if count >= maxReceivedImages {
		return fmt.Sprintf("【系统】已达到图片接收上限(%d张)，该图片被拒绝。\n\n请尽快输入和图片相关的问题。",
			maxReceivedImages)
	}

	url := msg.PicURL
	_ = store.AppendReceivedImageUrl(user, url)
	go func() {
		err := logic.CalAndStoreImageTokens(url)
		if err != nil {
			errorx.RecordError("CalAndStoreImageTokens() failed", err)
		}
	}()

	if count == maxReceivedImages-1 {
		reply = fmt.Sprintf("【系统】已接收%d张图片(上限)，请尽快输入和图片相关的问题。",
			count+1)
	} else {
		reply = fmt.Sprintf("【系统】已接收%d张图片(上限%d张），请尽快输入和图片相关的问题或者继续发送图片。",
			count+1, maxReceivedImages)
	}
	return
}
