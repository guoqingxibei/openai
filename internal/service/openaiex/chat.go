package openaiex

import "C"
import (
	"context"
	"errors"
	openai "github.com/sashabaranov/go-openai"
	"io"
	"log"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/util"
)

const CurrentModel = openai.GPT3Dot5Turbo

var ohmygptClient *openai.Client
var sbClient *openai.Client
var api2dClient *openai.Client
var ctx = context.Background()

func init() {
	ohmygptClient = createClientWithVendor(constant.Ohmygpt)
	sbClient = createClientWithVendor(constant.OpenaiSb)
	api2dClient = createClientWithVendor(constant.OpenaiApi2d)
}

func createClientWithVendor(aiVendor string) *openai.Client {
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

func createClient(key string, baseURL string) *openai.Client {
	var defaultConfig = openai.DefaultConfig(key)
	defaultConfig.BaseURL = baseURL + "/v1"
	return openai.NewClientWithConfig(defaultConfig)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func getClient(vendor string) *openai.Client {
	if vendor == constant.Ohmygpt {
		return ohmygptClient
	}
	if vendor == constant.OpenaiSb {
		return sbClient
	}
	return api2dClient
}

func CreateChatStream(
	messages []openai.ChatCompletionMessage,
	mode string,
	aiVendor string,
	processWord func(string2 string),
) (reply string, err error) {
	model := openai.GPT3Dot5Turbo
	if mode == constant.GPT4 {
		model = openai.GPT4
	}
	tokenCount := util.CalTokenCount4Messages(messages, CurrentModel)
	req := openai.ChatCompletionRequest{
		Model:     model,
		Messages:  messages,
		Stream:    true,
		MaxTokens: min(4000-tokenCount, 2000),
	}
	client := getClient(aiVendor)
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		log.Println("client.CreateChatCompletionStream() failed", err)
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return reply, nil
		}

		if err != nil {
			log.Println("stream.Recv() failed", err)
			return "", err
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			reply += content
			processWord(content)
		}
	}
}
