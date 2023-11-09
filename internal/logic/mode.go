package logic

import (
	"fmt"
	"openai/internal/constant"
)

func GetTimesPerQuestion(mode string) (times int) {
	switch mode {
	case constant.GPT3:
		times = constant.TimesPerQuestionGPT3
	case constant.GPT4:
		times = constant.TimesPerQuestionGPT4
	case constant.Draw:
		times = constant.TimesPerQuestionDraw
	case constant.Translate:
		times = constant.TimesPerQuestionTranslation
	}
	return
}

func GetModeDesc(mode string) (desc string) {
	switch mode {
	case constant.GPT3:
		fallthrough
	case constant.GPT4:
		desc = fmt.Sprintf("当前模式是%s，每次对话消耗次数%d。", GetModeName(mode), GetTimesPerQuestion(mode))
	case constant.Draw:
		desc = fmt.Sprintf("当前模式是%s，每次绘画消耗次数%d。", GetModeName(mode), GetTimesPerQuestion(mode))
	case constant.Translate:
		desc = fmt.Sprintf("当前模式是%s，每次翻译消耗次数%d。", GetModeName(mode), GetTimesPerQuestion(mode))
	}
	return
}

func GetModeName(mode string) (name string) {
	switch mode {
	case constant.GPT3:
		name = "GPT-3对话"
	case constant.GPT4:
		name = "GPT-4对话"
	case constant.Draw:
		name = "AI绘画"
	case constant.Translate:
		name = "中英互译"
	}
	return
}
