package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/robfig/cron"
	"github.com/silenceper/wechat/v2/officialaccount/material"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"net/url"
	"openai/internal/model"
	"openai/internal/service/errorx"
	"openai/internal/service/ohmygpt"
	"openai/internal/service/wechat"
	"openai/internal/store"
	"openai/internal/util"
	"path/filepath"
	"strings"
	"time"
)

const (
	imageDir = "midjourney-images"
)

var ctx = context.Background()

func init() {
	if !util.AccountIsUncle() || !util.EnvIsProd() {
		c1 := cron.New()
		// Execute once every ten seconds
		err := c1.AddFunc("*/10 * * * * *", func() {
			checkPendingTasks()
		})
		if err != nil {
			errorx.RecordError("AddFunc() failed", err)
			return
		}
		c1.Start()
	}
}

func SubmitDrawTask(prompt string, user string, mode string) string {
	taskIds, _ := store.GetPendingTaskIdsForUser(user)
	if len(taskIds) > 0 {
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		return "抱歉，你之前的画图任务仍在进行中，请稍后再提交新的任务。"
	}

	failureReply := "画图任务提交失败，请稍后重试，本次任务不会消耗次数。"
	taskResp, err := ohmygpt.SubmitDrawTask(prompt)
	if err != nil {
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		errorx.RecordError("ohmygpt.SubmitDrawTask() failed", err)
		return failureReply
	}

	if taskResp.StatusCode != 200 {
		if taskResp.Message != "" {
			failureReply += fmt.Sprintf("\n\n失败原因是「%s」", taskResp.Message)
		}
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		return failureReply
	}

	taskId := taskResp.Data
	onTaskCreated(user, taskId)
	return "画图任务已成功提交，作品将在1分钟后奉上！敬请期待..."
}

func checkPendingTasks() {
	taskIds, _ := store.GetPendingTaskIds()
	if len(taskIds) == 0 {
		return
	}

	for _, taskId := range taskIds {
		taskId := taskId
		go func() {
			err := checkTask(taskId)
			if err != nil {
				errorx.RecordError(fmt.Sprintf("checkTask(%d)", taskId), err)
			}
		}()
	}
}

func checkTask(taskId int) error {
	locker := store.GetLocker()
	lock, err := locker.Obtain(ctx, buildTaskLockKey(taskId), time.Minute*5, nil)
	defer func() {
		lock.Release(ctx)
		if !errors.Is(err, redislock.ErrNotObtained) {
			log.Printf("[task %d] Released task lock", taskId)
		}
	}()
	if errors.Is(err, redislock.ErrNotObtained) {
		return nil
	}

	log.Printf("[task %d] Obtained task lock, continue to check", taskId)
	statusResp, err := ohmygpt.GetTaskStatus(taskId)
	if err != nil {
		return err
	}

	if statusResp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("GetTaskStatus(%d) failed: status code is %d, error message is 「%s」",
			taskId,
			statusResp.StatusCode,
			statusResp.Message,
		))
	}

	user, _ := store.GetUserByTaskId(taskId)
	data := statusResp.Data
	status := data.Status
	action := data.Action
	log.Printf("[task %d] Status is %s, action is %s, user is %s", taskId, status, action, user)
	if status == ohmygpt.StatusSuccess {
		if action == ohmygpt.ActionImagine {
			log.Printf("[task %d] Executing UPSCALE actions...", taskId)
			for _, action := range data.Actions {
				customId := action.CustomId
				if strings.Contains(customId, "upsample::1") || strings.Contains(customId, "upsample::2") {
					subtaskId, _ := store.GetSubtaskId(taskId, customId)
					if subtaskId != 0 {
						log.Printf("[task %d] Skipped submitted subtask with customId: %s", taskId, customId)
						continue
					}

					subtaskId, err = ohmygpt.ExecuteAction(taskId, customId)
					if err != nil {
						return err
					}

					onSubtaskCreated(user, subtaskId, taskId, customId, data.FinishTime)
				}
			}
			onTaskFinished(user, taskId)
			log.Printf("[task %d] All eligible subtasks are submitted, removed this task", taskId)
			return nil
		}

		if action == ohmygpt.ActionUpscale {
			log.Printf("[task %d] Downloading image...", taskId)
			imageUrl := data.ImageDcUrl
			imageName, err := extractImageName(imageUrl)
			if err != nil {
				return err
			}

			filename := imageDir + "/" + imageName
			err = util.DownloadFile(imageUrl, filename)
			if err != nil {
				return err
			}

			log.Printf("[task %d] Uploading image to wechat...", taskId)
			media, err := wechat.GetAccount().GetMaterial().MediaUpload(material.MediaTypeImage, filename)
			if err != nil {
				return err
			}

			log.Printf("[task %d] Sending image to user...", taskId)
			err = wechat.GetAccount().
				GetCustomerMessageManager().Send(message.NewCustomerImgMessage(user, media.MediaID))
			if err != nil {
				return err
			}

			onTaskFinished(user, taskId)
			log.Printf("[task %d] Sent upscaled image to user, removed this task", taskId)
			return nil
		}
	}

	if status == ohmygpt.StatusFailure {
		if action == ohmygpt.ActionImagine {
			_, err := ohmygpt.SubmitDrawTask(data.Prompt)
			if err != nil {
				return err
			}

			onTaskFinished(user, taskId)
			onTaskCreated(user, taskId)
			return nil
		}

		if action == ohmygpt.ActionUpscale {
			task, _ := store.GetTask(taskId)
			subtaskId, err := ohmygpt.ExecuteAction(task.ParentTaskId, task.CustomId)
			if err != nil {
				return err
			}

			onTaskFinished(user, taskId)
			onSubtaskCreated(user, subtaskId, task.ParentTaskId, task.CustomId, task.ParentTaskFinishTime)
			return nil
		}
	}

	log.Printf("[task %d] Skipped", taskId)
	return nil
}

func extractImageName(imageUrl string) (imageName string, err error) {
	parsedURL, err := url.Parse(imageUrl)
	if err != nil {
		return
	}

	imageName = filepath.Base(parsedURL.Path)
	return
}

func buildTaskLockKey(taskId int) string {
	return fmt.Sprintf("task-id:%d:lock", taskId)
}

func onSubtaskCreated(user string, subtaskId int, taskId int, customId string, parentTaskFinishTime time.Time) {
	subtask := model.Task{
		TaskId:               subtaskId,
		ParentTaskId:         taskId,
		ParentTaskFinishTime: parentTaskFinishTime,
		CustomId:             customId,
	}
	_ = store.SetTask(subtaskId, subtask)
	_ = store.SetSubtaskId(taskId, customId, subtaskId)
	_ = store.AppendPendingTaskId(subtaskId)
	_ = store.AppendPendingTaskIdsForUser(user, subtaskId)
	_ = store.SetUserForTaskId(subtaskId, user)
}

func onTaskFinished(user string, taskId int) {
	_ = store.RemovePendingTaskId(taskId)
	_ = store.RemovePendingTaskIdForUser(user, taskId)
}

func onTaskCreated(user string, taskId int) {
	_ = store.AppendPendingTaskId(taskId)
	_ = store.AppendPendingTaskIdsForUser(user, taskId)
	_ = store.SetUserForTaskId(taskId, user)
}
