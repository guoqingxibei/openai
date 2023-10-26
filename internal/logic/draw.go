package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/robfig/cron"
	"github.com/silenceper/wechat/v2/officialaccount/material"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"golang.org/x/sync/errgroup"
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

const enableHighQualityImage = false

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
	if !util.IsEnglishSentence(prompt) {
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		return "由于MidJourney对非英文的支持非常有限，所以draw模式下暂时仅支持英文输入。"
	}

	taskIds, _ := store.GetPendingTaskIdsForUser(user)
	if len(taskIds) > 0 {
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		return "你仍有进行中的画图任务，请稍后提交新的任务。"
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
				errorx.RecordError(fmt.Sprintf("checkTask(%d) failed", taskId), err)
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
	if time.Now().After(data.StartTime.Add(time.Minute * 30)) {
		onTaskFinished(user, taskId)
		log.Printf("[task %d] Abandoned this task due to timeout", taskId)
		return errors.New("abandoned this task due to timeout")
	}

	if status == ohmygpt.StatusSuccess {
		if action == ohmygpt.ActionImagine {
			if !enableHighQualityImage {
				log.Printf("[task %d] Downloading image...", taskId)
				fileName, err := downloadImage(data.ImageDcUrl)
				if err != nil {
					return err
				}

				log.Printf("[task %d] Spliting images...", taskId)
				splitImages, err := util.SplitImage(fileName)
				if err != nil {
					return err
				}

				g := new(errgroup.Group)
				for _, splitImage := range splitImages {
					splitImage := splitImage
					g.Go(func() error {
						log.Printf("[task %d] Sending image to user...", taskId)
						return sendSplitImageToUser(splitImage, user)
					})
				}
				err = g.Wait()
				if err != nil {
					return err
				}

				onTaskFinished(user, taskId)
				return nil
			}

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
			image, err := downloadImage(data.ImageDcUrl)

			log.Printf("[task %d] Sending image to user...", taskId)
			err = sendImageToUser(image, user)
			if err != nil {
				return err
			}

			onTaskFinished(user, taskId)
			log.Printf("[task %d] Sent upscaled image to user, removed this task", taskId)
			return nil
		}
	}

	if status == ohmygpt.StatusFailure {
		reply := fmt.Sprintf("抱歉，任务执行失败，请稍后重试。失败原因是「%s」", data.FailReason)
		err = wechat.GetAccount().
			GetCustomerMessageManager().Send(message.NewCustomerTextMessage(user, reply))
		if err != nil {
			return err
		}

		onTaskFinished(user, taskId)
		log.Printf("[task %d] Abandoned this task due to failure, failure reason is 「%s」", taskId, data.FailReason)
		return nil

		if action == ohmygpt.ActionImagine {
			taskResp, err := ohmygpt.SubmitDrawTask(data.Prompt)
			if err != nil {
				return err
			}

			onTaskFinished(user, taskId)
			onTaskCreated(user, taskResp.Data)
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

func sendImageToUser(image string, user string) error {
	media, err := wechat.GetAccount().GetMaterial().MediaUpload(material.MediaTypeImage, image)
	if err != nil {
		return err
	}

	err = wechat.GetAccount().
		GetCustomerMessageManager().Send(message.NewCustomerImgMessage(user, media.MediaID))
	if err != nil {
		return err
	}
	return nil
}

func downloadImage(imageUrl string) (fileName string, err error) {
	imageName, err := extractImageName(imageUrl)
	if err != nil {
		return
	}

	fileName = imageDir + "/" + imageName
	err = util.DownloadFile(imageUrl, fileName)
	if err != nil {
		return
	}

	return
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

func sendSplitImageToUser(splitImage string, user string) error {
	sent, _ := store.GetImageSent(splitImage)
	if sent {
		return nil
	}

	log.Println(splitImage)
	err := sendImageToUser(splitImage, user)
	if err != nil {
		return err
	}

	_ = store.SetImageSent(splitImage)
	return nil
}
