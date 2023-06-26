package util

import (
	"encoding/json"
	"fmt"
	_openai "github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)

var enc, _ = tokenizer.Get(tokenizer.Cl100kBase)

func StringifyMessages(messages []_openai.ChatCompletionMessage) (string, error) {
	bytes, err := json.Marshal(messages)
	if err != nil {
		return "", nil
	}
	return string(bytes), nil
}

func ParseMessages(messagesStr string) ([]_openai.ChatCompletionMessage, error) {
	var messages []_openai.ChatCompletionMessage
	err := json.Unmarshal([]byte(messagesStr), &messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// Refer to https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
func getTokenCount(str string) int {
	tokenIds, _, _ := enc.Encode(str)
	return len(tokenIds)
}

func CalTokenCount4Messages(messages []_openai.ChatCompletionMessage, model string) int {
	var tokensPerMessage, tokensPerName int
	switch model {
	case _openai.GPT3Dot5Turbo:
		return CalTokenCount4Messages(messages, _openai.GPT3Dot5Turbo0301)
	case _openai.GPT4:
		return CalTokenCount4Messages(messages, _openai.GPT40314)
	case _openai.GPT3Dot5Turbo0301:
		tokensPerMessage = 4
		tokensPerName = -1
	case _openai.GPT40314:
		tokensPerMessage = 3
		tokensPerName = 1
	default:
		panic(fmt.Sprintf("num_tokens_from_messages() is not implemented for model %s.", model))
	}

	// Get message token count
	var numTokens int
	for _, message := range messages {
		numTokens += tokensPerMessage
		numTokens += getTokenCount(message.Role)
		numTokens += getTokenCount(message.Content)
		if message.Name != "" {
			numTokens += getTokenCount(message.Name) + tokensPerName
		}
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens
}

func RotateMessages(messages []_openai.ChatCompletionMessage, model string) ([]_openai.ChatCompletionMessage, error) {
	if model != _openai.GPT3Dot5Turbo {
		panic(fmt.Sprintf("RotateMessages() is not implemented for model %s.", model))
	}

	tokenCount := CalTokenCount4Messages(messages, model)
	for tokenCount > 2000 {
		messages = messages[1:]
		tokenCount = CalTokenCount4Messages(messages, model)
	}
	return messages, nil
}

func AppendAssistantMessage(messages []_openai.ChatCompletionMessage, answer string) []_openai.ChatCompletionMessage {
	messages = append(messages, _openai.ChatCompletionMessage{
		Role:    _openai.ChatMessageRoleAssistant,
		Content: answer,
	})
	return messages
}
