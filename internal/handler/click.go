package handler

import (
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"openai/internal/constant"
)

func onClick(msg *message.MixMessage) (reply *message.Reply) {
	log.Printf("%s clicked the button 「%s」", msg.FromUserName, msg.EventKey)
	switch msg.EventKey {
	case constant.GPT3:
		fallthrough
	case constant.GPT4:
		fallthrough
	case constant.Draw:
		reply = switchMode(msg.EventKey, msg)
	case clear:
		reply = clearHistory(msg)
	case help:
		reply = showUsage(msg)
	case invite:
		reply = getInvitationCode(msg)
	case donate:
		fallthrough
	case group:
		fallthrough
	case contact:
		reply = showImage(msg.EventKey)
	}
	return reply
}
