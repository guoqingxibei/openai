package logic

import (
	"github.com/sashabaranov/go-openai"
	"log/slog"
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

var aiVendors = []string{constant.Ohmygpt, constant.Ohmygpt, constant.Ohmygpt}

func CreateChatStreamEx(
	user string,
	msgId int64,
	question string,
	imageUrls []string,
	isVoice bool,
	mode string,
	reachMaxLengthChan chan<- bool,
) (fullReply string, fullReasoningReply string) {
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

	replyLen := 0
	chanSent := false
	reasoningDone := false
	for attemptNumber, vendor := range aiVendors {
		_ = store.DelReplyChunks(msgId)
		_ = store.AppendReplyChunk(msgId, startMark)

		if mode == constant.DeepSeekR1 {
			_ = store.DelReasoningReplyChunks(msgId)
			_ = store.AppendReasoningReplyChunk(msgId, startMark)
		}

		if isVoice {
			_ = store.AppendReplyChunk(msgId, "「"+question+"」\n\n")
		}

		// only ohmygpt support /gs feature
		if strings.HasPrefix(question, "/gs") {
			vendor = constant.Ohmygpt
		}
		nonSpaceStarts := false
		fullReply, fullReasoningReply, err = openaiex.CreateChatStream(
			messages,
			mode,
			maxOutputTokens,
			vendor,
			attemptNumber,
			func(word string) {
				if mode == constant.DeepSeekR1 && !reasoningDone {
					reasoningDone = true
					_ = store.AppendReasoningReplyChunk(msgId, endMark)
				}

				if !nonSpaceStarts {
					word = strings.TrimLeft(word, " \t\n")
					if word == "" {
						return
					}
				}

				nonSpaceStarts = true
				_ = store.AppendReplyChunk(msgId, word)
				replyLen += util.GetVisualLength(word)
				if !chanSent && replyLen > constant.MaxVisualLengthOfReply {
					reachMaxLengthChan <- true
					chanSent = true
				}
			},
			func(reasoningWord string) {
				_ = store.AppendReasoningReplyChunk(msgId, reasoningWord)
				replyLen += util.GetVisualLength(reasoningWord)
				if !chanSent && replyLen > constant.MaxVisualLengthOfReply {
					reachMaxLengthChan <- true
					chanSent = true
				}
			},
		)
		if err == nil {
			break
		}
		slog.Error("openaiex.CreateChatStream() failed",
			"msgId", msgId, "vendor", vendor, "mode", mode, "error", err)
	}
	if err != nil {
		return
	}

	_ = store.AppendReplyChunk(msgId, endMark)
	if mode != constant.Translate {
		messages = util.AppendAssistantMessage(messages, fullReply)
		_ = store.SetMessages(user, messages)
		_ = store.DelReceivedImageUrls(user)
	}
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
	reply := ""
	hasReasoning, _ := store.ReasoningReplyChunksExists(msgId)
	if hasReasoning {
		reasoningChunks, _ := store.GetReasoningReplyChunks(msgId, 1, -1)
		if len(reasoningChunks) <= 0 {
			return "", false
		}

		reply += "【开始思考...】\n"
		reasoningReachEnd := reasoningChunks[len(reasoningChunks)-1] == endMark
		if !reasoningReachEnd {
			reply += strings.Join(reasoningChunks, "")
			return reply, false
		}

		reply += strings.Join(reasoningChunks[:len(reasoningChunks)-1], "")
		if !strings.HasSuffix(reply, "\n") {
			reply += "\n"
		}
		reply += "【思考结束!】\n\n---\n"
	}

	chunks, _ := store.GetReplyChunks(msgId, 1, -1)
	if len(chunks) <= 0 {
		return reply, false
	}

	reachEnd := chunks[len(chunks)-1] == endMark
	if reachEnd {
		chunks = chunks[:len(chunks)-1]
	}
	reply += strings.Join(chunks, "")
	return reply, reachEnd
}

func FetchingReply(msgId int64, sendSegment func(segment string)) {
	var startIndex int64 = 1
	var reasoningStartIndex int64 = 1
	fetchTimes := 0
	hasReasoning, _ := store.ReasoningReplyChunksExists(msgId)
	if hasReasoning {
		sendSegment("【开始思考...】\n")
		reasoningEndWithNewLine := false
		for {
			fetchTimes++
			if fetchTimes > maxFetchTimes {
				break
			}

			reasoningChunks, _ := store.GetReasoningReplyChunks(msgId, reasoningStartIndex, -1)
			length := len(reasoningChunks)
			if length >= 1 {
				reasoningReachEnd := reasoningChunks[length-1] == endMark
				if reasoningReachEnd {
					reasoningChunks = reasoningChunks[:length-1]
				}
				segment := strings.Join(reasoningChunks, "")
				reasoningStartIndex += int64(length)
				if segment != "" {
					sendSegment(segment)
					reasoningEndWithNewLine = strings.HasSuffix(segment, "\n")
				}
				if reasoningReachEnd {
					break
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
		reasoningEndNote := "【思考结束!】\n\n---\n"
		if !reasoningEndWithNewLine {
			reasoningEndNote = "\n" + reasoningEndNote
		}
		sendSegment(reasoningEndNote)
	}

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
