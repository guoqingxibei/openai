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

func buildTaskIdUserKey(taskId int) string {
	return fmt.Sprintf("task-id:%d:user", taskId)
}

func SetUserForTaskId(taskId int, user string) (err error) {
	return client.Set(ctx, buildTaskIdUserKey(taskId), user, WEEK).Err()
}

func GetUserByTaskId(taskId int) (user string, err error) {
	return client.Get(ctx, buildTaskIdUserKey(taskId)).Result()
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
	return client.Set(ctx, buildSubtaskKey(taskId, customId), subtaskId, WEEK).Err()
}

func buildUserPendingTaskIdsKey(user string) string {
	return fmt.Sprintf("user:%s:pending-task-ids", user)
}

func AppendPendingTaskIdsForUser(user string, taskId int) (err error) {
	err = client.SAdd(ctx, buildUserPendingTaskIdsKey(user), taskId).Err()
	if err != nil {
		return
	}

	err = client.Expire(ctx, buildUserPendingTaskIdsKey(user), time.Minute*2).Err()
	return
}

func RemovePendingTaskIdForUser(user string, taskId int) (err error) {
	return client.SRem(ctx, buildUserPendingTaskIdsKey(user), taskId).Err()
}

func GetPendingTaskIdsForUser(user string) (ids []int, err error) {
	idStrs, err := client.SMembers(ctx, buildUserPendingTaskIdsKey(user)).Result()
	for _, idStr := range idStrs {
		id, _ := strconv.Atoi(idStr)
		ids = append(ids, id)
	}
	return
}
