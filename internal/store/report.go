package store

import (
	"encoding/json"
	"fmt"
	"openai/internal/model"
)

func buildActiveUsersKey(day string) string {
	return fmt.Sprintf("day:%s:active-users", day)
}

func appendActiveUser(day string, user string) error {
	err := client.SAdd(ctx, buildActiveUsersKey(day), user).Err()
	if err != nil {
		return err
	}

	return client.Expire(ctx, buildActiveUsersKey(day), WEEK).Err()
}

func GetActiveUsers(day string, useUncleDb bool) ([]string, error) {
	myClient := client
	if useUncleDb {
		myClient = uncleClient
	}

	return myClient.SMembers(ctx, buildActiveUsersKey(day)).Result()
}

func buildConversationsKey(user string, day string) string {
	return fmt.Sprintf("user:%s:day:%s:conversations", user, day)
}

func AppendConversation(user string, day string, conv model.Conversation) error {
	err := appendActiveUser(day, user)
	if err != nil {
		return err
	}

	convBytes, _ := json.Marshal(conv)
	err = client.RPush(ctx, buildConversationsKey(user, day), string(convBytes)).Err()
	if err != nil {
		return err
	}

	return client.Expire(ctx, buildConversationsKey(user, day), WEEK).Err()
}

func GetConversations(user string, day string, useUncleDb bool) ([]model.Conversation, error) {
	myClient := client
	if useUncleDb {
		myClient = uncleClient
	}

	var convs []model.Conversation
	convStrs, err := myClient.LRange(ctx, buildConversationsKey(user, day), 0, -1).Result()
	if err != nil {
		return convs, err
	}

	for _, convStr := range convStrs {
		var conv model.Conversation
		_ = json.Unmarshal([]byte(convStr), &conv)
		convs = append(convs, conv)
	}
	return convs, err
}
