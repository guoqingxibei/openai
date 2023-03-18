package util

import (
	"encoding/json"
	"openai/internal/service/openai"
)

func StringifyMessages(messages []openai.Message) (string, error) {
	bytes, err := json.Marshal(messages)
	if err != nil {
		return "", nil
	}
	return string(bytes), nil
}

func ParseMessages(messagesStr string) ([]openai.Message, error) {
	var messages []openai.Message
	err := json.Unmarshal([]byte(messagesStr), &messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
