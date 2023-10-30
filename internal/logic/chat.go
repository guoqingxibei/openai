package logic

import (
	_openai "github.com/sashabaranov/go-openai"
	"openai/internal/constant"
	"openai/internal/service/openai"
	"openai/internal/store"
	"openai/internal/util"
	"strings"
)

const (
	StartMark = "[START]"
	EndMark   = "[END]"
)

func ChatCompletionStream(
	aiVendor string, userName string, msgId int64,
	question string, isVoice bool, mode string,
) error {
	_ = store.AppendReplyChunk(msgId, StartMark)
	messages, err := store.GetMessages(userName)
	if err != nil {
		return err
	}

	messages = append(messages, _openai.ChatCompletionMessage{
		Role:    _openai.ChatMessageRoleUser,
		Content: question,
	})
	messages, err = util.RotateMessages(messages, openai.CurrentModel)
	if err != nil {
		return err
	}

	if isVoice {
		_ = store.AppendReplyChunk(msgId, "「"+question+"」\n\n")
	}
	var answer string
	openai.ChatCompletionsStream(
		aiVendor,
		mode,
		messages,
		func(word string) bool {
			answer += word
			_ = store.AppendReplyChunk(msgId, word)
			return true
		},
		func() {
			_ = store.AppendReplyChunk(msgId, EndMark)
			messages = util.AppendAssistantMessage(messages, answer)
			_ = store.SetMessages(userName, messages)
		},
		func(_err error) {
			_ = store.AppendReplyChunk(msgId, constant.TryAgain)
			_ = store.AppendReplyChunk(msgId, EndMark)
			err = _err
		},
	)
	return err
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
