package model

import "time"

type Task struct {
	TaskId               int       `json:"taskId"`
	ParentTaskId         int       `json:"parent_task_id"`
	ParentTaskFinishTime time.Time `json:"parent_task_finish_time"`
	CustomId             string    `json:"custom_id"`
}
