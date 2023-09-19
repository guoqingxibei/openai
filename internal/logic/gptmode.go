package logic

import "openai/internal/constant"

func GetTimesPerQuestion(gptMode string) int {
	times := constant.TimesPerQuestionGPT3
	if gptMode == constant.GPT4 {
		times = constant.TimesPerQuestionGPT4
	}
	return times
}
