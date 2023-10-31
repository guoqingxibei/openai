package logic

import (
	_openai "github.com/sashabaranov/go-openai"
	"log"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/service/openaiex"
	"openai/internal/store"
	"openai/internal/util"
	"strings"
)

const (
	StartMark = "[START]"
	EndMark   = "[END]"
)

var aiVendors = []string{constant.Ohmygpt, constant.OpenaiApi2d, constant.OpenaiSb}

func CreateChatStreamEx(
	user string,
	msgId int64,
	question string,
	isVoice bool,
	mode string,
) {
	messages, err := buildMessages(user, msgId, question)
	if err != nil {
		onFailure(user, msgId, mode, err)
		return
	}

	reply := ""
	for _, vendor := range aiVendors {
		_ = store.DelReplyChunks(msgId)
		_ = store.AppendReplyChunk(msgId, StartMark)
		if isVoice {
			_ = store.AppendReplyChunk(msgId, "「"+question+"」\n\n")
		}
		reply, err = openaiex.CreateChatStream(messages, mode, vendor,
			func(word string) {
				_ = store.AppendReplyChunk(msgId, word)
			},
		)
		if err == nil {
			break
		}
		log.Printf("openaiex.CreateChatStream(%d, %s) failed: %v", msgId, vendor, err)
	}

	if err != nil {
		onFailure(user, msgId, mode, err)
		return
	}

	_ = store.AppendReplyChunk(msgId, EndMark)
	messages = util.AppendAssistantMessage(messages, reply)
	_ = store.SetMessages(user, messages)
}

func onFailure(user string, msgId int64, mode string, err error) {
	AddPaidBalance(user, GetTimesPerQuestion(mode))
	_ = store.DelReplyChunks(msgId)
	_ = store.AppendReplyChunk(msgId, StartMark)
	_ = store.AppendReplyChunk(msgId, constant.TryAgain)
	_ = store.AppendReplyChunk(msgId, EndMark)
	errorx.RecordError("CreateChatStreamEx() failed", err)
}

func buildMessages(user string, msgId int64, question string) (messages []_openai.ChatCompletionMessage, err error) {
	messages, err = store.GetMessages(user)
	if err != nil {
		return
	}

	messages = append(messages, _openai.ChatCompletionMessage{
		Role:    _openai.ChatMessageRoleUser,
		Content: question,
	})
	messages, err = util.RotateMessages(messages, openaiex.CurrentModel)
	return
}

func FetchReply(msgId int64) (string, bool) {
	chunks, _ := store.GetReplyChunks(msgId, 1, -1)
	if len(chunks) <= 0 {
		return "", false
	}

	reachEnd := chunks[len(chunks)-1] == EndMark
	if reachEnd {
		chunks = chunks[:len(chunks)-1]
	}
	reply := strings.Join(chunks, "")
	return reply, reachEnd
}
