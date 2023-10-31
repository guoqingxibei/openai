package logic

import (
	"github.com/sashabaranov/go-openai"
	"log"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/service/openaiex"
	"openai/internal/store"
	"openai/internal/util"
	"strings"
	"time"
)

const (
	startMark = "[START]"
	endMark   = "[END]"

	maxFetchTimes = 6000
)

var aiVendors = []string{constant.Ohmygpt, constant.OpenaiApi2d, constant.OpenaiSb}

func CreateChatStreamEx(
	user string,
	msgId int64,
	question string,
	isVoice bool,
	mode string,
) {
	model := getModel(mode)
	messages, err := buildMessages(user, question, model)
	if err != nil {
		onFailure(user, msgId, mode, err)
		return
	}

	reply := ""
	for _, vendor := range aiVendors {
		_ = store.DelReplyChunks(msgId)
		_ = store.AppendReplyChunk(msgId, startMark)
		if isVoice {
			_ = store.AppendReplyChunk(msgId, "「"+question+"」\n\n")
		}
		reply, err = openaiex.CreateChatStream(messages, model, vendor,
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

	_ = store.AppendReplyChunk(msgId, endMark)
	messages = util.AppendAssistantMessage(messages, reply)
	_ = store.SetMessages(user, messages)
}

func getModel(mode string) string {
	model := openai.GPT3Dot5Turbo
	if mode == constant.GPT4 {
		model = openai.GPT4
	}
	return model
}

func onFailure(user string, msgId int64, mode string, err error) {
	AddPaidBalance(user, GetTimesPerQuestion(mode))
	_ = store.DelReplyChunks(msgId)
	_ = store.AppendReplyChunk(msgId, startMark)
	_ = store.AppendReplyChunk(msgId, constant.TryAgain)
	_ = store.AppendReplyChunk(msgId, endMark)
	errorx.RecordError("CreateChatStreamEx() failed", err)
}

func buildMessages(user string, question string, model string) (
	messages []openai.ChatCompletionMessage,
	err error,
) {
	messages, err = store.GetMessages(user)
	if err != nil {
		return
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})
	messages, err = util.RotateMessages(messages, model)
	return
}

func FetchReply(msgId int64) (string, bool) {
	chunks, _ := store.GetReplyChunks(msgId, 1, -1)
	if len(chunks) <= 0 {
		return "", false
	}

	reachEnd := chunks[len(chunks)-1] == endMark
	if reachEnd {
		chunks = chunks[:len(chunks)-1]
	}
	reply := strings.Join(chunks, "")
	return reply, reachEnd
}

func FetchingReply(msgId int64, sendSegment func(segment string)) {
	var startIndex int64 = 1
	fetchTimes := 0
	for {
		fetchTimes++
		if fetchTimes > maxFetchTimes {
			break
		}

		chunks, _ := store.GetReplyChunks(msgId, startIndex, -1)
		length := len(chunks)
		if length >= 1 {
			reachEnd := chunks[length-1] == endMark
			if reachEnd {
				chunks = chunks[:length-1]
			}
			segment := strings.Join(chunks, "")
			sendSegment(segment)
			startIndex += int64(length)
			if reachEnd {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
