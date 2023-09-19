package openai

import "C"
import (
	"context"
	"errors"
	_openai "github.com/sashabaranov/go-openai"
	"io"
	"log"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/util"
)

const CurrentModel = _openai.GPT3Dot5Turbo

var sbClient *_openai.Client
var api2dClient *_openai.Client
var ctx = context.Background()

func init() {
	sbClient = createClientWithVendor(constant.OpenaiSb)
	api2dClient = createClientWithVendor(constant.OpenaiApi2d)
}

func createClientWithVendor(aiVendor string) *_openai.Client {
	if aiVendor == constant.OpenaiSb {
		return createClient(config.C.OpenaiSb.Key, config.C.OpenaiSb.BaseURL)
	}
	return createClient(config.C.OpenaiApi2d.Key, config.C.OpenaiApi2d.BaseURL)
}

func createClient(key string, baseURL string) *_openai.Client {
	var defaultConfig = _openai.DefaultConfig(key)
	defaultConfig.BaseURL = baseURL
	return _openai.NewClientWithConfig(defaultConfig)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func getClient(vendor string) *_openai.Client {
	if vendor == constant.OpenaiSb {
		return sbClient
	}
	return api2dClient
}

func ChatCompletionsStream(
	aiVendor string,
	gptMode string,
	messages []_openai.ChatCompletionMessage,
	processWord func(word string) bool,
	done func(),
	errorHandler func(err error)) {
	model := _openai.GPT3Dot5Turbo
	if gptMode == constant.GPT4 {
		model = _openai.GPT4
	}
	tokenCount := util.CalTokenCount4Messages(messages, CurrentModel)
	req := _openai.ChatCompletionRequest{
		Model:     model,
		Messages:  messages,
		Stream:    true,
		MaxTokens: min(4000-tokenCount, 2000),
	}
	client := getClient(aiVendor)
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		errorHandler(err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			done()
			return
		}

		if err != nil {
			errorHandler(err)
			return
		}

		if len(response.Choices) > 0 {
			ok := processWord(response.Choices[0].Delta.Content)
			if !ok {
				break
			}
		} else {
			log.Printf("error: reponse.Choices is empty, current response is %v", response)
		}
	}
}
