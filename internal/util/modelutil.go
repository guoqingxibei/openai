package util

import (
	"github.com/sashabaranov/go-openai"
	"openai/internal/constant"
)

func GetModelByMode(mode string) (model string) {
	switch mode {
	case constant.GPT3:
		model = openai.GPT4oMini
	case constant.GPT4:
		model = openai.GPT4o20240806
	case constant.Translate:
		model = openai.GPT4oMini
	}
	return
}
