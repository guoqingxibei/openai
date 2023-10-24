package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/go-http-utils/headers"
	"github.com/robfig/cron"
	"github.com/silenceper/wechat/v2/officialaccount/material"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/service/wechat"
	"openai/internal/store"
	"openai/internal/util"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type commonResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type taskResponse struct {
	commonResponse
	Data int `json:"data"` // taskId
}

type statusResponse struct {
	commonResponse
	Data struct {
		Status     string `json:"status"`
		FailReason string `json:"failReason"`
		ImageDcUrl string `json:"imageDcUrl"`
		Action     string `json:"action"`
		Actions    []struct {
			CustomId string `json:"customId"`
			Label    string `json:"label"`
		}
	} `json:"data"`
}

const (
	imageDir       = "midjourney-images"
	modeMidJourney = "midjourney"
	typeNormal     = "NORMAL"
	actionUpscale  = "UPSCALE"
	actionImagine  = "IMAGINE"
	statusSuccess  = "SUCCESS"
)

var ctx = context.Background()

func init() {
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

func SubmitDrawTask(prompt string, user string, mode string) string {
	start := time.Now()
	imagineUrl := config.C.Ohmygpt.BaseURL + "/api/v1/ai/draw/mj/imagine"

	params := url.Values{}
	params.Set("model", modeMidJourney)
	params.Set("prompt", prompt)
	params.Set("type", typeNormal)
	data := params.Encode()
	payload := strings.NewReader(data)

	req, err := http.NewRequest(http.MethodPost, imagineUrl, payload)
	failureReply := "画图任务提交失败，请稍后重试，本次任务不会消耗次数。"
	if err != nil {
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		errorx.RecordError("http.NewRequest() failed", err)
		return failureReply
	}

	req.Header.Add(headers.Authorization, constant.AuthorizationPrefixBearer+config.C.Ohmygpt.Key)
	req.Header.Add(headers.ContentType, constant.ContentTypeFormURLEncoded)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		errorx.RecordError("client.Do() failed", err)
		return failureReply
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		errorx.RecordError("ioutil.ReadAll() failed", err)
		return failureReply
	}

	var taskResp taskResponse
	_ = json.Unmarshal(body, &taskResp)
	log.Printf("[SubmitDrawTaskAPI] Duration: %dms, prompt: 「%s」,response: 「%s」",
		int(time.Since(start).Milliseconds()),
		prompt,
		util.EscapeNewline(string(body)),
	)
	if taskResp.StatusCode != 200 {
		if taskResp.Message != "" {
			failureReply += fmt.Sprintf("\n\n失败原因是「%s」", taskResp.Message)
		}
		AddPaidBalance(user, GetTimesPerQuestion(mode))
		return failureReply
	}

	taskId := taskResp.Data
	_ = store.AppendPendingTaskId(taskId)
	_ = store.SetUserForTaskId(taskId, user)
	return "画图任务已成功提交，作品稍后奉上！敬请期待..."
}

func getTaskStatus(taskId int) (statusResp *statusResponse, err error) {
	start := time.Now()
	getStatusUrl := config.C.Ohmygpt.BaseURL + "/api/v1/ai/draw/mj/query"

	params := url.Values{}
	params.Set("model", modeMidJourney)
	params.Set("taskId", strconv.Itoa(taskId))
	data := params.Encode()
	payload := strings.NewReader(data)

	req, err := http.NewRequest(http.MethodPost, getStatusUrl, payload)
	if err != nil {
		return
	}

	req.Header.Add(headers.Authorization, constant.AuthorizationPrefixBearer+config.C.Ohmygpt.Key)
	req.Header.Add(headers.ContentType, constant.ContentTypeFormURLEncoded)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	statusResp = &statusResponse{}
	_ = json.Unmarshal(body, &statusResp)
	log.Printf("[GetTaskStatusAPI] Duration: %dms, taskId: %d, response: 「%s」",
		int(time.Since(start).Milliseconds()),
		taskId,
		util.EscapeNewline(string(body)),
	)
	if statusResp.StatusCode != 200 {
		err = errors.New(
			fmt.Sprintf("GetTaskStatus API failed with status code %d, response is 「%s」",
				statusResp.StatusCode,
				body,
			),
		)
	}
	return
}

func downloadImage(imageUrl string, fileName string) (err error) {
	start := time.Now()
	response, err := http.Get(imageUrl)
	if err != nil {
		return
	}
	defer response.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return
	}

	log.Printf("[DownloadImageAPI] Duration: %dms, fileName: %s, imageUrl: %s",
		int(time.Since(start).Milliseconds()),
		fileName,
		imageUrl,
	)
	return
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

func buildTaskLockKey(taskId int) string {
	return fmt.Sprintf("task-id:%d:lock", taskId)
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
	statusResp, err := getTaskStatus(taskId)
	if err != nil {
		return err
	}

	data := statusResp.Data
	status := data.Status
	action := data.Action
	log.Printf("[task %d] Status is %s, action is %s", taskId, status, action)
	if status == statusSuccess {
		if action == actionImagine {
			log.Printf("[task %d] Executing UPSCALE actions...", taskId)
			for _, action := range data.Actions {
				customId := action.CustomId
				if strings.Contains(customId, "upsample") {
					subtaskId, _ := store.GetSubtaskId(taskId, customId)
					if subtaskId != 0 {
						log.Printf("[task %d] Skipped submitted subtask with customId: %s", taskId, customId)
						continue
					}

					subtaskId, err = executeAction(taskId, customId)
					if err != nil {
						return err
					}

					_ = store.SetSubtaskId(taskId, customId, subtaskId)
					_ = store.AppendPendingTaskId(subtaskId)
					user, _ := store.GetUserOfTaskId(taskId)
					_ = store.SetUserForTaskId(subtaskId, user)
				}
			}
			_ = store.RemovePendingTaskId(taskId)
			log.Printf("[task %d] All eligible subtasks are submitted, removed this task", taskId)
			return nil
		}

		if action == actionUpscale {
			log.Printf("[task %d] Downloading image...", taskId)
			imageUrl := data.ImageDcUrl
			imageName, err := extractImageName(imageUrl)
			if err != nil {
				return err
			}

			filename := imageDir + "/" + imageName
			err = downloadImage(imageUrl, filename)
			if err != nil {
				return err
			}

			log.Printf("[task %d] Uploading image to wechat...", taskId)
			media, err := wechat.GetAccount().GetMaterial().MediaUpload(material.MediaTypeImage, filename)
			if err != nil {
				return err
			}

			log.Printf("[task %d] Sending image to user...", taskId)
			user, _ := store.GetUserOfTaskId(taskId)
			err = wechat.GetAccount().
				GetCustomerMessageManager().Send(message.NewCustomerImgMessage(user, media.MediaID))
			if err != nil {
				return err
			}

			log.Printf("[task %d] Sent upscaled image to user, removed this task", taskId)
			_ = store.RemovePendingTaskId(taskId)
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

func executeAction(taskId int, customId string) (subTaskId int, err error) {
	start := time.Now()
	actionUrl := config.C.Ohmygpt.BaseURL + "/api/v1/ai/draw/mj/action"

	params := url.Values{}
	params.Set("model", modeMidJourney)
	params.Set("taskId", strconv.Itoa(taskId))
	params.Set("customId", customId)
	params.Set("type", typeNormal)

	data := params.Encode()
	payload := strings.NewReader(data)

	req, err := http.NewRequest(http.MethodPost, actionUrl, payload)
	if err != nil {
		errorx.RecordError("http.NewRequest() failed", err)
		return
	}

	req.Header.Add(headers.Authorization, constant.AuthorizationPrefixBearer+config.C.Ohmygpt.Key)
	req.Header.Add(headers.ContentType, constant.ContentTypeFormURLEncoded)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		errorx.RecordError("client.Do() failed", err)
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		errorx.RecordError("ioutil.ReadAll() failed", err)
		return
	}

	var taskResp taskResponse
	_ = json.Unmarshal(body, &taskResp)
	log.Printf("[ExecuteActionAPI] Duration: %dms, taskId: %d, customeId: %s, response: 「%s」",
		int(time.Since(start).Milliseconds()),
		taskId,
		customId,
		util.EscapeNewline(string(body)),
	)
	if taskResp.StatusCode != 200 {
		errMsg := "ExecuteActionAPI failed"
		if taskResp.Message != "" {
			errMsg = fmt.Sprintf("%s, reason is 「%s」", errMsg, taskResp.Message)
		}
		err = errors.New(errMsg)
		return
	}

	subTaskId = taskResp.Data
	return
}
