package openaiex

import (
	"context"
	"github.com/sashabaranov/go-openai"
	"log"
	"openai/internal/util"
	"strings"
	"time"
)

func VoiceToText(voiceFile string, vendor string) (text string, err error) {
	start := time.Now()
	defer func() {
		log.Printf("[VoiceToText] Duration: %dms, voiceFile: 「%s」,text: 「%s」",
			int(time.Since(start).Milliseconds()),
			voiceFile,
			util.EscapeNewline(text),
		)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	client := getClient(vendor)
	response, err := client.CreateTranscription(ctx, openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: voiceFile,
		Prompt:   "以下是普通话的句子。",
		Format:   openai.AudioResponseFormatText,
	})
	text = strings.TrimSpace(response.Text)
	return
}
