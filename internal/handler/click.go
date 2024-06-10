package handler

import (
	"fmt"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log/slog"
	"openai/internal/constant"
	"openai/internal/logic"
)

func onClick(msg *message.MixMessage) (reply *message.Reply) {
	slog.Info(fmt.Sprintf("%s clicked the button 「%s」", msg.FromUserName, msg.EventKey))
	switch msg.EventKey {
	case constant.GPT3:
		fallthrough
	case constant.GPT4:
		fallthrough
	case constant.Draw:
		fallthrough
	case constant.TTS:
		fallthrough
	case constant.Translate:
		reply = switchMode(msg.EventKey, msg)
	case clear:
		reply = clearHistory(msg)
	case help:
		reply = logic.ShowUsage(msg)
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
