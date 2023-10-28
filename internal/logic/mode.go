package logic

import (
	"fmt"
	"openai/internal/constant"
)

func GetTimesPerQuestion(mode string) (times int) {
	if mode == constant.GPT3 {
		return constant.TimesPerQuestionGPT3
	}

	if mode == constant.GPT4 {
		return constant.TimesPerQuestionGPT4
	}
	return constant.TimesPerQuestionDraw
}

func GetModeDesc(mode string) string {
	if mode == constant.Draw {
		return fmt.Sprintf("当前模式是%s，每次画图消耗次数%d。", GetModeName(mode), GetTimesPerQuestion(mode))
	}

	return fmt.Sprintf("当前模式是%s，每次对话消耗次数%d。", GetModeName(mode), GetTimesPerQuestion(mode))
}

func GetModeName(mode string) string {
	if mode == constant.Draw {
		return "AI画图"
	}

	if mode == constant.GPT3 {
		return "GPT-3对话"
	}

	return "GPT-4对话"
}
