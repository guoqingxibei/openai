package openaiex

import "C"
import (
	"context"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/util"
	"runtime/debug"
	"time"
)

var ohmygptClient *openai.Client
var sbClient *openai.Client
var gptApiUsClient *openai.Client
var openaiClient *openai.Client

func init() {
	ohmygptClient = createClientWithVendor(constant.Ohmygpt)
	sbClient = createClientWithVendor(constant.OpenaiSb)
	gptApiUsClient = createClientWithVendor(constant.GptApiUs)
	openaiClient = createClientWithVendor(constant.Openai)
}

func createClientWithVendor(aiVendor string) *openai.Client {
	if aiVendor == constant.Ohmygpt {
		baseUrl := config.C.Ohmygpt.BaseURL
		if config.C.Ohmygpt.UseAzure {
			baseUrl += "/azure"
		}
		return createClient(config.C.Ohmygpt.Key, baseUrl)
	}

	if aiVendor == constant.GptApiUs {
		return createClient(config.C.GptApiUs.Key, config.C.GptApiUs.BaseURL)
	}

	if aiVendor == constant.Openai {
		return createClient(config.C.Openai.Key, config.C.Openai.BaseURL)
	}

	return createClient(config.C.OpenaiSb.Key, config.C.OpenaiSb.BaseURL)
}

func createClient(key string, baseURL string) *openai.Client {
	var defaultConfig = openai.DefaultConfig(key)
	defaultConfig.BaseURL = baseURL + "/v1"
	return openai.NewClientWithConfig(defaultConfig)
}

func getClient(vendor string) *openai.Client {
	if vendor == constant.Ohmygpt {
		return ohmygptClient
	}

	if vendor == constant.GptApiUs {
		return gptApiUsClient
	}

	if vendor == constant.Openai {
		return openaiClient
	}

	return sbClient
}

func CreateChatStream(
	messages []openai.ChatCompletionMessage,
	mode string,
	maxTokens int,
	aiVendor string,
	attemptNumber int,
	processWord func(string),
) (reply string, _err error) {
	req := openai.ChatCompletionRequest{
		Model:     util.GetModelByMode(mode),
		Messages:  messages,
		MaxTokens: maxTokens,
		Stream:    true,
		StreamOptions: &openai.StreamOptions{
			IncludeUsage: true,
		},
	}
	client := getClient(aiVendor)

	doneChan := make(chan bool, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicMsg := fmt.Sprintf("%v\n%s", r, debug.Stack())
				errorx.RecordError("failed due to a panic", errors.New(panicMsg))
			}
		}()

		stream, err := client.CreateChatCompletionStream(ctx, req)
		if err != nil {
			_err = fmt.Errorf("client.CreateChatCompletionStream() failed: %w", err)
			doneChan <- true
			return
		}
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				_err = nil
				doneChan <- true
				return
			}

			if err != nil {
				_err = fmt.Errorf("stream.Recv() failed: %w", err)
				doneChan <- true
				return
			}

			if len(response.Choices) > 0 {
				content := response.Choices[0].Delta.Content
				reply += content
				processWord(content)
			} else {
				log.Println(fmt.Sprintf("[TokensUsage] %+v", response.Usage))
			}
		}
	}()

	// fail if the first word didn't reach in specified time
	timeout := getTimeout(mode, attemptNumber)
	select {
	case <-time.After(time.Second * time.Duration(timeout)):
		if reply == "" {
			cancel()
			errMsg := fmt.Sprintf("not yet start responding in %d seconds", timeout)
			return "", errors.New(errMsg)
		}

		<-doneChan
	case <-doneChan:
	}
	return
}

func getTimeout(mode string, attemptNumber int) (timeout int) {
	base := 5
	step := 3
	if mode == constant.GPT4 {
		base = 10
		step = 5
	}
	timeout = base + attemptNumber*step
	return
}

func TransToEng(original string, vendor string) (trans string, err error) {
	start := time.Now()
	defer func() {
		log.Printf("[TransToEngAPI] Duration: %dms, original: 「%s」,trans: 「%s」",
			int(time.Since(start).Milliseconds()),
			original,
			util.EscapeNewline(trans),
		)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client := getClient(vendor)
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:            openai.GPT3Dot5Turbo,
			Messages:         util.BuildTransMessages(original, constant.English),
			MaxTokens:        1000,
			FrequencyPenalty: 1,
			PresencePenalty:  1,
			Temperature:      0,
			TopP:             1,
		},
	)
	if err != nil {
		return
	}

	if len(resp.Choices) <= 0 {
		err = errors.New("no available choice")
		return
	}

	trans = resp.Choices[0].Message.Content
	return
}
