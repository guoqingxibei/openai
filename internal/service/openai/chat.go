package openai

import "C"
import (
	"context"
	"errors"
	_openai "github.com/sashabaranov/go-openai"
	"io"
	"log"
	"openai/internal/config"
	"openai/internal/util"
)

const CurrentModel = _openai.GPT3Dot5Turbo

var client *_openai.Client
var ctx = context.Background()

func init() {
	var defaultConfig = _openai.DefaultConfig(config.C.OpenAI.Key)
	defaultConfig.BaseURL = config.C.OpenAI.BaseURL
	client = _openai.NewClientWithConfig(defaultConfig)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func ChatCompletionsStream(
	messages []_openai.ChatCompletionMessage,
	processWord func(word string) bool,
	done func(),
	errorHandler func(err error)) {
	tokenCount := util.CalTokenCount4Messages(messages, CurrentModel)
	req := _openai.ChatCompletionRequest{
		Model:     _openai.GPT3Dot5Turbo,
		Messages:  messages,
		Stream:    true,
		MaxTokens: min(4000-tokenCount, 2000),
	}
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
