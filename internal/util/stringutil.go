package util

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"strings"
	"unicode"
)

func EscapeNewline(originStr string) string {
	return strings.ReplaceAll(originStr, "\n", `\n`)
}

func EscapeHtmlTags(origin string) string {
	replaced := strings.ReplaceAll(origin, "<", "&lt;")
	replaced = strings.ReplaceAll(replaced, ">", "&gt;")
	return replaced
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

func MarkdownToHtml(md string) string {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock | parser.HardLineBreak
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))

	// create HTML renderer with extensions
	htmlFlags := html.CompletePage
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return string(markdown.Render(doc, renderer))
}

func GetRuneLength(s string) int {
	return len([]rune(s))
}

// GetVisualLength
// 1 English char is 1 visual unit
// 1 Chinese char is 2 visual units
func GetVisualLength(s string) int {
	length := 0
	for _, r := range s {
		length += getVisualLengthOfChar(r)
	}
	return length
}

func TruncateReplyVisually(s string, visualLength int) string {
	truncatedReply := ""
	length := 0
	for _, r := range s {
		if length > visualLength {
			break
		}
		length += getVisualLengthOfChar(r)
		truncatedReply += string(r)
	}
	return truncatedReply
}

func getVisualLengthOfChar(r rune) int {
	if r > unicode.MaxASCII {
		return 2
	}

	return 1
}
