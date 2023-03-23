package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"openai/internal/config"
	"openai/internal/service/wechat"
	"strings"
	"sync/atomic"
	"time"
)

const chatUrl = "https://api.openai.com/v1/chat/completions"

var totalTokens int64

type request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type response struct {
	ID    string `json:"id"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	// Object  string                 `json:"object"`
	// Created int                    `json:"created"`
	// Model   string                 `json:"model"`
	Choices []choiceItem `json:"choices"`
	// Usage   map[string]interface{} `json:"usage"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

type choiceItem struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

// ChatCompletions https://beta.openai.com/docs/api-reference/making-requests
func ChatCompletions(messages []Message, shortMsgId string, inMsg *wechat.Msg) (string, error) {
	start := time.Now()
	var r request
	r.Model = "gpt-3.5-turbo"
	r.Messages = messages

	bs, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: time.Second * 300}
	req, _ := http.NewRequest("POST", chatUrl, bytes.NewReader(bs))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+config.C.OpenAI.Key)

	// 设置代理
	if config.C.Http.Proxy != "" {
		proxyURL, _ := url.Parse(config.C.Http.Proxy)
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data response
	json.Unmarshal(body, &data)
	statusCode := resp.StatusCode
	lastQuestion := messages[len(messages)-1].Content
	if statusCode >= 200 && statusCode < 300 && len(data.Choices) > 0 {
		atomic.AddInt64(&totalTokens, int64(data.Usage.TotalTokens))
		lastAnswer := strings.TrimSpace(data.Choices[0].Message.Content)
		log.Printf("User: %s, message ID: %d, short message ID: %s, duration: %ds, "+
			"request tokens：%d, response tokens: %d, question:「%s」, answer:「%s」",
			inMsg.FromUserName,
			inMsg.MsgId,
			shortMsgId,
			int(time.Since(start).Seconds()),
			data.Usage.PromptTokens,
			data.Usage.CompletionTokens,
			escapeNewline(lastQuestion),
			escapeNewline(lastAnswer),
		)

		return lastAnswer, nil
	}

	errorMsg := data.Error.Message
	log.Printf("User: %s, message ID: %d, short message ID: %s, duration: %ds, "+
		"question:「%s」, error:「%s」",
		inMsg.FromUserName,
		inMsg.MsgId,
		shortMsgId,
		int(time.Since(start).Seconds()),
		escapeNewline(lastQuestion),
		escapeNewline(errorMsg),
	)
	return "", errors.New(fmt.Sprintf("Error %d: %s", statusCode, errorMsg))
}

func escapeNewline(originStr string) string {
	return strings.ReplaceAll(originStr, "\n", `\n`)
}
