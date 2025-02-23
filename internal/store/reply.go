package store

import "strconv"

func AppendReplyChunk(msgId int64, chunk string) error {
	err := client.RPush(ctx, buildReplyChunksKey(msgId), chunk).Err()
	if err != nil {
		return err
	}
	err = client.Expire(ctx, buildReplyChunksKey(msgId), WEEK).Err()
	return err
}

func ReplyChunksExists(msgId int64) (bool, error) {
	code, err := client.Exists(ctx, buildReplyChunksKey(msgId)).Result()
	return code == 1, err
}

func GetReplyChunks(msgId int64, from int64, to int64) ([]string, error) {
	return client.LRange(ctx, buildReplyChunksKey(msgId), from, to).Result()
}

func DelReplyChunks(msgId int64) error {
	return client.Del(ctx, buildReplyChunksKey(msgId)).Err()
}

func buildReplyChunksKey(msgId int64) string {
	return "msg-id:" + strconv.FormatInt(msgId, 10) + ":reply-chunks"
}

func AppendReasoningReplyChunk(msgId int64, chunk string) error {
	err := client.RPush(ctx, buildReasoningReplyChunksKey(msgId), chunk).Err()
	if err != nil {
		return err
	}
	err = client.Expire(ctx, buildReasoningReplyChunksKey(msgId), WEEK).Err()
	return err
}

func ReasoningReplyChunksExists(msgId int64) (bool, error) {
	code, err := client.Exists(ctx, buildReasoningReplyChunksKey(msgId)).Result()
	return code == 1, err
}

func GetReasoningReplyChunks(msgId int64, from int64, to int64) ([]string, error) {
	return client.LRange(ctx, buildReasoningReplyChunksKey(msgId), from, to).Result()
}

func DelReasoningReplyChunks(msgId int64) error {
	return client.Del(ctx, buildReasoningReplyChunksKey(msgId)).Err()
}

func buildReasoningReplyChunksKey(msgId int64) string {
	return "msg-id:" + strconv.FormatInt(msgId, 10) + ":reasoning-reply-chunks"
}
