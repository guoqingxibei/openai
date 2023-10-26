package store

import (
	"github.com/redis/go-redis/v9"
	_openai "github.com/sashabaranov/go-openai"
	"openai/internal/util"
	"time"
)

func SetMessages(toUserName string, messages []_openai.ChatCompletionMessage) error {
	newRoundsStr, err := util.StringifyMessages(messages)
	if err != nil {
		return err
	}
	return client.Set(ctx, buildMessagesKey(toUserName), newRoundsStr, time.Minute*5).Err()
}

func GetMessages(toUserName string) ([]_openai.ChatCompletionMessage, error) {
	var messages []_openai.ChatCompletionMessage
	messagesStr, err := client.Get(ctx, buildMessagesKey(toUserName)).Result()
	if err != nil {
		if err == redis.Nil {
			return messages, nil
		}
		return nil, err
	}
	messages, err = util.ParseMessages(messagesStr)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func DelMessages(toUserName string) error {
	return client.Del(ctx, buildMessagesKey(toUserName)).Err()
}

func buildMessagesKey(toUserName string) string {
	return "user:" + toUserName + ":messages"
}
