package logic

import "openai/internal/constant"

func GetTimesPerQuestion(mode string) (times int) {
	if mode == constant.GPT3 {
		return constant.TimesPerQuestionGPT3
	}

	if mode == constant.GPT4 {
		return constant.TimesPerQuestionGPT4
	}
	return constant.TimesPerQuestionDraw
}
