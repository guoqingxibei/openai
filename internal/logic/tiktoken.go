package logic

import (
	"errors"
	"fmt"
	"github.com/pkoukk/tiktoken-go"
	"github.com/redis/go-redis/v9"
	"github.com/sashabaranov/go-openai"
	"log/slog"
	"openai/internal/store"
	"strings"
)

// calTokensForMessages
// OpenAI Cookbook: https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
func calTokensForMessages(messages []openai.ChatCompletionMessage, model string) (numTokens int, err error) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return
	}

	var tokensPerMessage, tokensPerName int
	switch model {
	case "gpt-3.5-turbo-0613",
		"gpt-3.5-turbo-16k-0613",
		"gpt-4-0314",
		"gpt-4-32k-0314",
		"gpt-4-0613",
		"gpt-4-32k-0613":
		tokensPerMessage = 3
		tokensPerName = 1
	case "gpt-3.5-turbo-0301":
		tokensPerMessage = 4 // every message follows <|start|>{role/name}\n{content}<|end|>\n
		tokensPerName = -1   // if there's a name, the role is omitted
	default:
		if strings.Contains(model, "gpt-3.5-turbo") {
			return calTokensForMessages(messages, "gpt-3.5-turbo-0613")
		} else if strings.Contains(model, "gpt-4") {
			return calTokensForMessages(messages, "gpt-4-0613")
		} else {
			err = fmt.Errorf("num_tokens_from_messages() is not implemented for model %s. "+
				"See https://github.com/openai/openai-python/blob/main/chatml.md "+
				"for information on how messages are converted to tokens", model)
			return
		}
	}

	for _, message := range messages {
		numTokens += tokensPerMessage
		// image input
		if len(message.MultiContent) > 0 {
			for _, part := range message.MultiContent {
				if part.Type == openai.ChatMessagePartTypeText {
					numTokens += len(tkm.Encode(part.Text, nil, nil)) + 6
				} else {
					url := part.ImageURL.URL
					tokens, err := store.GetImageTokens(url)
					if errors.Is(err, redis.Nil) {
						slog.Info("No tokens found for image", "url", url)
						tokens = 85
					}
					numTokens += tokens + 8
				}
			}
		} else {
			numTokens += len(tkm.Encode(message.Content, nil, nil))
		}
		numTokens += len(tkm.Encode(message.Role, nil, nil))
		numTokens += len(tkm.Encode(message.Name, nil, nil))
		if message.Name != "" {
			numTokens += tokensPerName
		}
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens, nil
}
