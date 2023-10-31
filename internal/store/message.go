package store

import (
	"github.com/redis/go-redis/v9"
	"github.com/sashabaranov/go-openai"
	"openai/internal/util"
	"time"
)

func SetMessages(user string, messages []openai.ChatCompletionMessage) error {
	messagesStr, err := util.StringifyMessages(messages)
	if err != nil {
		return err
	}
	return client.Set(ctx, buildMessagesKey(user), messagesStr, time.Minute*5).Err()
}

func GetMessages(user string) ([]openai.ChatCompletionMessage, error) {
	var messages []openai.ChatCompletionMessage
	messagesStr, err := client.Get(ctx, buildMessagesKey(user)).Result()
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

func DelMessages(user string) error {
	return client.Del(ctx, buildMessagesKey(user)).Err()
}

func buildMessagesKey(user string) string {
	return "user:" + user + ":messages"
}
