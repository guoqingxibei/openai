package store

import (
	"fmt"
	"strconv"
	"time"
)

func getPendingDrawTaskIdsKey() string {
	return "pending-draw-task-ids"
}

func GetPendingTaskIds() (ids []int, err error) {
	idStrs, err := client.SMembers(ctx, getPendingDrawTaskIdsKey()).Result()
	for _, idStr := range idStrs {
		id, _ := strconv.Atoi(idStr)
		ids = append(ids, id)
	}
	return
}

func AppendPendingTaskId(taskId int) (err error) {
	return client.SAdd(ctx, getPendingDrawTaskIdsKey(), taskId).Err()
}

func RemovePendingTaskId(taskId int) (err error) {
	return client.SRem(ctx, getPendingDrawTaskIdsKey(), taskId).Err()
}

func buildUserKeyWithTaskId(taskId int) string {
	return fmt.Sprintf("task-id:%d:user", taskId)
}

func SetUserForTaskId(taskId int, user string) (err error) {
	return client.Set(ctx, buildUserKeyWithTaskId(taskId), user, time.Hour*24*7).Err()
}

func GetUserOfTaskId(taskId int) (user string, err error) {
	return client.Get(ctx, buildUserKeyWithTaskId(taskId)).Result()
}

func buildSubtaskKey(parentTaskId int, customId string) string {
	return fmt.Sprintf("task-id:%d:custom-id:%s:sub-task-id", parentTaskId, customId)
}

func GetSubtaskId(taskId int, customId string) (subtaskId int, err error) {
	subtaskIdStr, err := client.Get(ctx, buildSubtaskKey(taskId, customId)).Result()
	if err != nil {
		return
	}

	subtaskId, err = strconv.Atoi(subtaskIdStr)
	return
}

func SetSubtaskId(taskId int, customId string, subtaskId int) (err error) {
	return client.Set(ctx, buildSubtaskKey(taskId, customId), subtaskId, time.Hour*24*7).Err()
}
