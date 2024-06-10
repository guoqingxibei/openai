package ohmygpt

import (
	"encoding/json"
	"fmt"
	"github.com/go-http-utils/headers"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/url"
	"openai/internal/config"
	"openai/internal/constant"
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
		SubmitTime time.Time `json:"submitTime"`
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
	TypeFast       = "FAST"

	StatusSuccess = "SUCCESS"
	StatusFailure = "FAILURE"
)

func SubmitDrawTask(prompt string) (response *TaskResponse, err error) {
	response = &TaskResponse{}

	start := time.Now()
	imagineUrl := config.C.Ohmygpt.BaseURL + "/api/v1/ai/draw/mj/imagine"

	params := url.Values{}
	params.Set("model", ModeMidJourney)
	params.Set("prompt", prompt)
	params.Set("type", TypeFast)
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
	slog.Info(fmt.Sprintf("[SubmitDrawTaskAPI] Duration: %dms, prompt: 「%s」,response: 「%s」",
		int(time.Since(start).Milliseconds()),
		prompt,
		util.EscapeNewline(string(body)),
	))
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
	slog.Info(fmt.Sprintf("[GetTaskStatusAPI] Duration: %dms, taskId: %d, response: 「%s」",
		int(time.Since(start).Milliseconds()),
		taskId,
		util.EscapeNewline(string(body)),
	))
	return
}
