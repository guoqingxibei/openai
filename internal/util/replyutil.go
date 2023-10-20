package util

import "github.com/silenceper/wechat/v2/officialaccount/message"

func BuildTextReply(reply string) *message.Reply {
	text := message.NewText(reply)
	return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
}

func BuildImageReply(mediaID string) *message.Reply {
	image := message.NewImage(mediaID)
	return &message.Reply{MsgType: message.MsgTypeImage, MsgData: image}
}
