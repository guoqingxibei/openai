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

func buildUserPendingTaskIdsKey(user string) string {
	return fmt.Sprintf("user:%s:pending-task-ids", user)
}

func AppendPendingTaskIdsForUser(user string, taskId int) (err error) {
	err = client.SAdd(ctx, buildUserPendingTaskIdsKey(user), taskId).Err()
	if err != nil {
		return
	}

	err = client.Expire(ctx, buildUserPendingTaskIdsKey(user), time.Minute*10).Err()
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

func buildImageSentKey(imageName string) string {
	return fmt.Sprintf("image:%s:sent", imageName)
}

func SetImageSent(imageName string) error {
	return client.Set(ctx, buildImageSentKey(imageName), strconv.FormatBool(true), WEEK).Err()
}

func GetImageSent(imageName string) (sent bool, err error) {
	sentStr, err := client.Get(ctx, buildImageSentKey(imageName)).Result()
	if err != nil {
		return
	}

	return sentStr == strconv.FormatBool(true), nil
}
