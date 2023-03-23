package util

import "strings"

func EscapeNewline(originStr string) string {
	return strings.ReplaceAll(originStr, "\n", `\n`)
}
