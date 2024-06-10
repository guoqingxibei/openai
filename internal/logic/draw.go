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
	"log/slog"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/service/ohmygpt"
	"openai/internal/service/wechat"
	"openai/internal/store"
	"openai/internal/util"
	"runtime/debug"
	"time"
)

const (
	mdImageDir = constant.Temp + "/midjourney-images"
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
		return "你仍有进行中的绘画任务，请稍后提交新的任务。"
	}

	replyPrefix := ""
	if !util.IsEnglishSentence(prompt) {
		trans, err := transToEngEx(prompt)
		if err != nil {
			AddPaidBalance(user, GetTimesPerQuestion(mode))
			errorx.RecordError("openaiex.transToEng() failed", err)
			return "翻译失败，请重试。"
		}
		prompt = trans
		replyPrefix = fmt.Sprintf("检测到非英文输入，系统自动翻译为「%s」\n\n", trans)
	}

	failureReply := replyPrefix + "绘画任务提交失败，请稍后重试，本次任务不会消耗次数。"
	taskResp, err := ohmygpt.SubmitDrawTask(prompt)
	if err != nil {
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		errorx.RecordError("ohmygpt.SubmitDrawTask() failed", err)
		return failureReply
	}

	if taskResp.StatusCode != 200 {
		if taskResp.Message != "" {
			failureReply += fmt.Sprintf("失败原因是「%s」", taskResp.Message)
		}
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		return failureReply
	}

	taskId := taskResp.Data
	onTaskCreated(user, taskId)
	return replyPrefix + "绘画任务已提交，作品将在2分钟后奉上！敬请期待..."
}

func checkPendingTasks() {
	taskIds, _ := store.GetPendingTaskIds()
	if len(taskIds) == 0 {
		return
	}

	for _, taskId := range taskIds {
		taskId := taskId
		go func() {
			defer func() {
				if r := recover(); r != nil {
					panicMsg := fmt.Sprintf("%v\n%s", r, debug.Stack())
					errorx.RecordError("failed due to a panic", errors.New(panicMsg))
				}
			}()

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
			slog.Info(fmt.Sprintf("[task %d] Released task lock", taskId))
		}
	}()
	if errors.Is(err, redislock.ErrNotObtained) {
		return nil
	}

	slog.Info(fmt.Sprintf("[task %d] Obtained task lock, continue to check", taskId))
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
	slog.Info(fmt.Sprintf("[task %d] Status is %s, action is %s, user is %s", taskId, status, action, user))
	if time.Now().After(data.SubmitTime.Add(time.Minute * 30)) {
		onTaskFinished(user, taskId, false)
		slog.Info(fmt.Sprintf("[task %d] Abandoned this task due to timeout", taskId))
		return errors.New("abandoned this task due to timeout")
	}

	if status == ohmygpt.StatusSuccess {
		slog.Info(fmt.Sprintf("[task %d] Downloading image...", taskId))
		filePath, err := util.DownloadFileInto(data.ImageDcUrl, mdImageDir)
		if err != nil {
			return err
		}

		slog.Info(fmt.Sprintf("[task %d] Spliting images...", taskId))
		splitImages, err := util.SplitImage(filePath)
		if err != nil {
			return err
		}

		g := new(errgroup.Group)
		for _, splitImage := range splitImages {
			splitImage := splitImage
			g.Go(func() error {
				slog.Info(fmt.Sprintf("[task %d] Sending image to user...", taskId))
				return sendSplitImageToUser(splitImage, user)
			})
		}
		err = g.Wait()
		if err != nil {
			return err
		}

		slog.Info(fmt.Sprintf("[task %d] Took %fs", taskId, time.Since(data.SubmitTime).Seconds()))
		onTaskFinished(user, taskId, true)
		return nil
	}

	if status == ohmygpt.StatusFailure {
		reply := fmt.Sprintf("抱歉，任务执行失败，请稍后重试。失败原因是「%s」", data.FailReason)
		err = wechat.GetAccount().
			GetCustomerMessageManager().Send(message.NewCustomerTextMessage(user, reply))
		if err != nil {
			return err
		}

		onTaskFinished(user, taskId, false)
		slog.Info(fmt.Sprintf("[task %d] Abandoned this task due to failure, failure reason is 「%s」", taskId, data.FailReason))
		return nil
	}

	slog.Info(fmt.Sprintf("[task %d] Skipped", taskId))
	return nil
}

func sendImageToUser(image string, user string) error {
	media, err := wechat.GetAccount().GetMaterial().MediaUpload(material.MediaTypeImage, image)
	if err != nil {
		return err
	}

	return wechat.GetAccount().
		GetCustomerMessageManager().Send(message.NewCustomerImgMessage(user, media.MediaID))
}

func buildTaskLockKey(taskId int) string {
	return fmt.Sprintf("task-id:%d:lock", taskId)
}

func onTaskCreated(user string, taskId int) {
	_ = store.AppendPendingTaskId(taskId)
	_ = store.AppendPendingTaskIdsForUser(user, taskId)
	_ = store.SetUserForTaskId(taskId, user)
}

func onTaskFinished(user string, taskId int, isSuccessful bool) {
	_ = store.RemovePendingTaskId(taskId)
	_ = store.RemovePendingTaskIdForUser(user, taskId)
	if !isSuccessful {
		AddPaidBalance(user, constant.TimesPerQuestionDraw)
	}
}

func sendSplitImageToUser(splitImage string, user string) error {
	sent, _ := store.GetImageSent(splitImage)
	if sent {
		return nil
	}

	err := sendImageToUser(splitImage, user)
	if err != nil {
		return err
	}

	_ = store.SetImageSent(splitImage)
	return nil
}
