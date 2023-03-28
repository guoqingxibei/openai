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
	"time"
)

const imageUrl = "https://api.openai.com/v1/images/generations"

type imageRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

type imageResponse struct {
	Items []item `json:"data"`
	Error struct {
		Message string `json:"message"`
	}
}

type item struct {
	Url string `json:"url"`
}

func GenerateImage(prompt string) (string, error) {
	start := time.Now()
	var r = imageRequest{
		Prompt: prompt,
		N:      1,
		Size:   "512x512",
	}
	bs, err := json.Marshal(r)
	if err != nil {
		log.Println("json.Marshal(r) failed", err)
		return "", errors.New(constant.TryAgain)
	}

	client := &http.Client{Timeout: time.Second * 300}
	req, _ := http.NewRequest("POST", imageUrl, bytes.NewReader(bs))
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
		log.Println("Failed to generate image", err)
		return "", errors.New(constant.TryAgain)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ioutil.ReadAll failed", err)
		return "", errors.New(constant.TryAgain)
	}

	var response imageResponse
	json.Unmarshal(body, &response)
	statusCode := resp.StatusCode
	if statusCode >= 200 && statusCode < 300 && len(response.Items) > 0 {
		url := response.Items[0].Url
		log.Printf("[ImageAPI] Duration: %dms, prompt:「%s」, image: %s",
			int(time.Since(start).Milliseconds()),
			prompt,
			url,
		)
		return url, nil
	}

	errorMsg := response.Error.Message
	log.Printf("[ImageAPI] Duration: %dms, prompt:「%s」, error: %s",
		int(time.Since(start).Milliseconds()),
		prompt,
		util.EscapeNewline(fmt.Sprintf("%d: %s", statusCode, errorMsg)),
	)
	return "", errors.New(errorMsg)
}
