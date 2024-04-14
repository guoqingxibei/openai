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
