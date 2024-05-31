package logic

import (
	"github.com/sashabaranov/go-openai"
	"log"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/service/openaiex"
	"openai/internal/store"
	"openai/internal/util"
	"slices"
	"strings"
	"time"
)

const (
	startMark = "[START]"
	endMark   = "[END]"

	maxFetchTimes = 6000

	maxInputTokens  = 4000
	maxOutputTokens = 4000 // ~2000 Chinese characters
)

var aiVendors = []string{constant.Openai, constant.Ohmygpt, constant.OpenaiSb}

func CreateChatStreamEx(
	user string,
	msgId int64,
	question string,
	imageUrls []string,
	isVoice bool,
	mode string,
) (fullReply string) {
	var err error = nil
	defer func() {
		if err != nil {
			onFailure(user, msgId, mode, err)
		}
	}()

	messages, err := buildMessages(user, question, imageUrls, mode)
	if err != nil {
		return
	}

	for attemptNumber, vendor := range aiVendors {
		_ = store.DelReplyChunks(msgId)
		_ = store.AppendReplyChunk(msgId, startMark)
		if isVoice {
			_ = store.AppendReplyChunk(msgId, "「"+question+"」\n\n")
		}

		// only ohmygpt support /gs feature
		if strings.HasPrefix(question, "/gs") {
			vendor = constant.Ohmygpt
		}
		fullReply, err = openaiex.CreateChatStream(
			messages,
			mode,
			maxOutputTokens,
			vendor,
			attemptNumber,
			func(word string) {
				_ = store.AppendReplyChunk(msgId, word)
			},
		)
		if err == nil {
			break
		}
		log.Printf("openaiex.CreateChatStream(%d, %s, %s) failed: %v", msgId, vendor, mode, err)
	}
	if err != nil {
		return
	}

	_ = store.AppendReplyChunk(msgId, endMark)
	messages = util.AppendAssistantMessage(messages, fullReply)
	_ = store.SetMessages(user, messages)
	_ = store.DelReceivedImageUrls(user)
	return
}

func onFailure(user string, msgId int64, mode string, err error) {
	urls, _ := store.GetReceivedImageUrls(user)
	if len(urls) > 0 { // when input image
		// add back count consumed by images
		AddPaidBalance(user, GetTimesPerQuestion(mode)*len(urls))
		_ = store.DelReceivedImageUrls(user)
	}

	AddPaidBalance(user, GetTimesPerQuestion(mode))
	_ = store.DelReplyChunks(msgId)
	_ = store.AppendReplyChunk(msgId, startMark)
	_ = store.AppendReplyChunk(msgId, constant.TryAgain)
	_ = store.AppendReplyChunk(msgId, endMark)
	errorx.RecordError("CreateChatStreamEx() failed", err)
}

func buildMessages(user string, question string, imageUrls []string, mode string) (
	messages []openai.ChatCompletionMessage,
	err error,
) {
	if mode == constant.Translate {
		targetLang := constant.English
		if util.IsEnglishSentence(question) {
			targetLang = constant.Chinese
		}
		messages = util.BuildTransMessages(question, targetLang)
		return
	}

	messages, err = store.GetMessages(user)
	if err != nil {
		return
	}

	var newMessage openai.ChatCompletionMessage
	if len(imageUrls) > 0 {
		textPart := openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeText,
			Text: question,
		}
		multiContent := []openai.ChatMessagePart{textPart}
		for _, url := range imageUrls {
			multiContent = append(multiContent, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    url,
					Detail: openai.ImageURLDetailAuto,
				},
			})
		}
		newMessage = openai.ChatCompletionMessage{
			Role:         openai.ChatMessageRoleUser,
			MultiContent: multiContent,
		}
	} else {
		newMessage = openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: question,
		}
	}

	// gpt-3 doesn't support vision grammar
	if mode == constant.GPT3 {
		messages = slices.DeleteFunc(messages, func(message openai.ChatCompletionMessage) bool {
			return len(message.MultiContent) > 0
		})
	}
	messages = append(messages, newMessage)
	messages, err = rotateMessages(messages, util.GetModelByMode(mode))
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

func rotateMessages(messages []openai.ChatCompletionMessage, model string) ([]openai.ChatCompletionMessage, error) {
	tokenCount, err := calTokensForMessages(messages, model)
	if err != nil {
		return nil, err
	}

	for tokenCount > maxInputTokens {
		// keep at least one message
		if len(messages) <= 1 {
			break
		}

		messages = messages[1:]
		tokenCount, err = calTokensForMessages(messages, model)
		if err != nil {
			return nil, err
		}
	}
	return messages, nil
}
