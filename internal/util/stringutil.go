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
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}
