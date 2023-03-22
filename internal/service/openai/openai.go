package openai

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"openai/internal/config"
	"strings"
	"sync/atomic"
	"time"
)

const (
	api          = "https://api.openai.com/v1/chat/completions"
	exchangeRate = 6.9
)

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

// Completions https://beta.openai.com/docs/api-reference/making-requests
func Completions(messages []Message, timeout time.Duration) (string, error) {
	start := time.Now()
	var r request
	r.Model = "gpt-3.5-turbo"
	r.Messages = messages

	bs, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: timeout}
	req, _ := http.NewRequest("POST", api, bytes.NewReader(bs))
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
	if len(data.Choices) > 0 {
		atomic.AddInt64(&totalTokens, int64(data.Usage.TotalTokens))

		reply := replyMsg(data.Choices[0].Message.Content)
		log.Printf("Duration: %ds，Request token：%d, Response token: %d\nQuestion: %s\nAnswer: %s\nRaw messages: %v",
			int(time.Since(start).Seconds()),
			data.Usage.PromptTokens,
			data.Usage.CompletionTokens,
			messages[len(messages)-1].Content,
			reply,
			messages,
		)

		return reply, nil
	}

	return data.Error.Message, nil
}

func replyMsg(reply string) string {
	idx := strings.Index(reply, "\n\n")
	if idx > -1 && reply[len(reply)-2] != '\n' {
		reply = reply[idx+2:]
	}
	start := 0
	for i, v := range reply {
		if !isSymbol(v) {
			start = i
			break
		}
	}

	return reply[start:]
}

var symbols = []rune{'\n', ' ', '，', '。', '？', '?', ',', '.', '!', '！', ':', '：'}

func isSymbol(w rune) bool {
	for _, v := range symbols {
		if v == w {
			return true
		}
	}
	return false
}
