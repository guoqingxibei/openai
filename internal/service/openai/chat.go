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
	"openai/internal/constant"
	"openai/internal/util"
	"strings"
	"sync/atomic"
	"time"
)

const chatUrl = "https://api.openai.com/v1/chat/completions"

var totalTokens int64

type request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature"`
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
func ChatCompletions(messages []Message) (string, error) {
	start := time.Now()
	r := request{
		Model:       "gpt-3.5-turbo",
		Temperature: 0.5,
		Messages:    messages,
	}

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
		log.Printf(
			"[CompletionAPI] Duration: %dms, question:「%s」, answer:「%s」",
			int(time.Since(start).Milliseconds()),
			util.EscapeNewline(lastQuestion),
			util.EscapeNewline(lastAnswer),
		)

		return lastAnswer, nil
	}

	errorMsg := data.Error.Message
	errorMsg = util.EscapeNewline(fmt.Sprintf("%d: %s", statusCode, errorMsg))
	log.Printf("[CompletionAPI] Duration: %dms, question:「%s」, error:「%s」",
		int(time.Since(start).Milliseconds()),
		util.EscapeNewline(lastQuestion),
		errorMsg,
	)
	return "", errors.New(errorMsg)
}

func StringifyMessages(messages []Message) (string, error) {
	bytes, err := json.Marshal(messages)
	if err != nil {
		return "", nil
	}
	return string(bytes), nil
}

func ParseMessages(messagesStr string) ([]Message, error) {
	var messages []Message
	err := json.Unmarshal([]byte(messagesStr), &messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func RotateMessages(messages []Message) ([]Message, error) {
	str, err := StringifyMessages(messages)
	for len(str) > 3000 {
		messages = messages[1:]
		str, err = StringifyMessages(messages)
		if err != nil {
			log.Println("stringifyMessages failed", err)
			return nil, err
		}
	}
	return messages, nil
}

func PrependSystemMessage(messages []Message) []Message {
	messages = append([]Message{
		{
			Role:    "system",
			Content: constant.ChatSystemMessage,
		},
	}, messages...)
	return messages
}

func RemoveSystemMessage(messages []Message) []Message {
	return messages[1:]
}

func AppendAssistantMessage(messages []Message, answer string) []Message {
	messages = append(messages, Message{
		Role:    "assistant",
		Content: answer,
	})
	return messages
}
