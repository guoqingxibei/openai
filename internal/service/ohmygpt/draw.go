package ohmygpt

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-http-utils/headers"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/util"
	"strconv"
	"strings"
	"time"
)

type commonResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

type TaskResponse struct {
	commonResponse
	Data int `json:"data"` // taskId
}

type StatusResponse struct {
	commonResponse
	Data struct {
		Status     string    `json:"status"`
		Prompt     string    `json:"prompt"`
		FinishTime time.Time `json:"finishTime"`
		FailReason string    `json:"failReason"`
		ImageDcUrl string    `json:"imageDcUrl"`
		Action     string    `json:"action"`
		Actions    []struct {
			CustomId string `json:"customId"`
			Label    string `json:"label"`
		}
	} `json:"data"`
}

const (
	ModeMidJourney = "midjourney"
	TypeNormal     = "NORMAL"
	ActionUpscale  = "UPSCALE"
	ActionImagine  = "IMAGINE"

	StatusSuccess = "SUCCESS"
	StatusFailure = "FAILURE"
)

func ExecuteAction(taskId int, customId string) (subTaskId int, err error) {
	start := time.Now()
	actionUrl := config.C.Ohmygpt.BaseURL + "/api/v1/ai/draw/mj/action"

	params := url.Values{}
	params.Set("model", ModeMidJourney)
	params.Set("taskId", strconv.Itoa(taskId))
	params.Set("customId", customId)
	params.Set("type", TypeNormal)

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

	var taskResp TaskResponse
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

func SubmitDrawTask(prompt string) (response *TaskResponse, err error) {
	response = &TaskResponse{}

	start := time.Now()
	imagineUrl := config.C.Ohmygpt.BaseURL + "/api/v1/ai/draw/mj/imagine"

	params := url.Values{}
	params.Set("model", ModeMidJourney)
	params.Set("prompt", prompt)
	params.Set("type", TypeNormal)
	data := params.Encode()
	payload := strings.NewReader(data)

	req, err := http.NewRequest(http.MethodPost, imagineUrl, payload)
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

	_ = json.Unmarshal(body, response)
	log.Printf("[SubmitDrawTaskAPI] Duration: %dms, prompt: 「%s」,response: 「%s」",
		int(time.Since(start).Milliseconds()),
		prompt,
		util.EscapeNewline(string(body)),
	)
	return
}

func GetTaskStatus(taskId int) (statusResp *StatusResponse, err error) {
	statusResp = &StatusResponse{}

	start := time.Now()
	getStatusUrl := config.C.Ohmygpt.BaseURL + "/api/v1/ai/draw/mj/query"

	params := url.Values{}
	params.Set("model", ModeMidJourney)
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

	_ = json.Unmarshal(body, &statusResp)
	log.Printf("[GetTaskStatusAPI] Duration: %dms, taskId: %d, response: 「%s」",
		int(time.Since(start).Milliseconds()),
		taskId,
		util.EscapeNewline(string(body)),
	)
	return
}
