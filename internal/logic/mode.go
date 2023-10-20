package logic

import "openai/internal/constant"

func GetTimesPerQuestion(mode string) int {
	times := constant.TimesPerQuestionGPT3
	if mode == constant.GPT4 {
		times = constant.TimesPerQuestionGPT4
	}
	return times
}
