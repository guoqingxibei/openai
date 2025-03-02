package logic

import (
	"fmt"
	"openai/internal/constant"
	"unicode/utf8"
)

func GetTimesPerQuestion(mode string) (times int) {
	switch mode {
	case constant.GPT3:
		times = constant.TimesPerQuestionGPT3
	case constant.GPT4:
		times = constant.TimesPerQuestionGPT4
	case constant.GPT4Dot5:
		times = constant.TimesPerQuestionGPT4Dot5
	case constant.DeepSeekR1:
		times = constant.TimesPerQuestionDeepSeekR1
	case constant.Draw:
		times = constant.TimesPerQuestionDraw
	case constant.Translate:
		times = constant.TimesPerQuestionTranslation
	}
	return
}

func calTimesForTTS(text string) int {
	return divideAndCeil(utf8.RuneCountInString(text), constant.CharCountPerTimeTTS)
}

func GetModeDesc(mode string) (desc string) {
	switch mode {
	case constant.GPT3:
		fallthrough
	case constant.GPT4:
		fallthrough
	case constant.GPT4Dot5:
		fallthrough
	case constant.DeepSeekR1:
		desc = fmt.Sprintf("当前模式是%s，每次对话消耗次数%d。", GetModeName(mode), GetTimesPerQuestion(mode))
	case constant.Draw:
		desc = fmt.Sprintf("当前模式是%s，每次绘画消耗次数%d。", GetModeName(mode), GetTimesPerQuestion(mode))
	case constant.TTS:
		desc = fmt.Sprintf("当前模式是%s，每%d字消耗次数1。", GetModeName(mode), constant.CharCountPerTimeTTS)
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
	case constant.GPT4Dot5:
		name = "GPT-4.5对话"
	case constant.DeepSeekR1:
		name = "DeepSeek-R1对话"
	case constant.Draw:
		name = "AI绘画"
	case constant.TTS:
		name = "文字转语音"
	case constant.Translate:
		name = "中英互译"
	}
	return
}

func divideAndCeil(num, divisor int) int {
	if divisor == 0 {
		panic("divisor cannot be zero")
	}

	result := num / divisor

	if num%divisor == 0 {
		return result
	}

	return result + 1
}
