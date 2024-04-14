package util

import (
	"strings"
	"unicode"
)

func EscapeNewline(originStr string) string {
	return strings.ReplaceAll(originStr, "\n", `\n`)
}

func IsEnglishSentence(sentence string) bool {
	for _, r := range sentence {
		if r > unicode.MaxASCII && !unicode.IsPunct(r) {
			return false
		}
	}
	return true
}

func TruncateString(origin string, maxLen int) string {
	runes := []rune(origin)
	if len(runes) > maxLen {
		halfLen := maxLen / 2
		return string(runes[:halfLen]) + "..." + string(runes[len(runes)-halfLen:])
	}

	return origin
}

func TruncateAndEscapeNewLine(originStr string, maxLen int) string {
	return EscapeNewline(TruncateString(originStr, maxLen))
}
