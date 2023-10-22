package openai

import "C"
import (
	"context"
	"errors"
	"fmt"
	_openai "github.com/sashabaranov/go-openai"
	"io"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/service/recorder"
	"openai/internal/util"
)

const CurrentModel = _openai.GPT3Dot5Turbo

var ohmygptClient *_openai.Client
var sbClient *_openai.Client
var api2dClient *_openai.Client
var ctx = context.Background()

func init() {
	ohmygptClient = createClientWithVendor(constant.Ohmygpt)
	sbClient = createClientWithVendor(constant.OpenaiSb)
	api2dClient = createClientWithVendor(constant.OpenaiApi2d)
}

func createClientWithVendor(aiVendor string) *_openai.Client {
	if aiVendor == constant.Ohmygpt {
		baseUrl := config.C.Ohmygpt.BaseURL
		if config.C.Ohmygpt.UseAzure {
			baseUrl += "/azure"
		}
		return createClient(config.C.Ohmygpt.Key, baseUrl)
	}
	if aiVendor == constant.OpenaiSb {
		return createClient(config.C.OpenaiSb.Key, config.C.OpenaiSb.BaseURL)
	}
	return createClient(config.C.OpenaiApi2d.Key, config.C.OpenaiApi2d.BaseURL)
}

func createClient(key string, baseURL string) *_openai.Client {
	var defaultConfig = _openai.DefaultConfig(key)
	defaultConfig.BaseURL = baseURL + "/v1"
	return _openai.NewClientWithConfig(defaultConfig)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func getClient(vendor string) *_openai.Client {
	if vendor == constant.Ohmygpt {
		return ohmygptClient
	}
	if vendor == constant.OpenaiSb {
		return sbClient
	}
	return api2dClient
}

func ChatCompletionsStream(
	aiVendor string,
	mode string,
	messages []_openai.ChatCompletionMessage,
	processWord func(word string) bool,
	done func(),
	errorHandler func(err error)) {
	model := _openai.GPT3Dot5Turbo
	if mode == constant.GPT4 {
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
			recorder.RecordError("Choices is empty", errors.New(fmt.Sprintf("response is [%s]", response)))
		}
	}
}
