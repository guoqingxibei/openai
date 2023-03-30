package openailogic

import (
	"openai/internal/constant"
	"openai/internal/service/baidu"
	"openai/internal/service/gptredis"
	"openai/internal/service/openai"
)

func ChatCompletion(userName string, question string) (string, error) {
	messages, err := gptredis.FetchMessages(userName)
	if err != nil {
		return "", err
	}
	messages = append(messages, openai.Message{
		Role:    "user",
		Content: question,
	})
	messages, err = openai.RotateMessages(messages)
	if err != nil {
		return "", err
	}
	messages = openai.PrependSystemMessage(messages)
	answer, err := openai.ChatCompletions(messages)
	if err != nil {
		return "", err
	}
	messages = openai.RemoveSystemMessage(messages)
	messages = openai.AppendAssistantMessage(messages, answer)
	err = gptredis.SetMessages(userName, messages)
	if err != nil {
		return "", err
	}
	passedCensor := baidu.Censor(answer)
	if !passedCensor {
		answer = constant.CensorWarning
	}
	return answer, err
}
