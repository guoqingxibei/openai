package openaiex

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"log/slog"
	"openai/internal/util"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func VoiceToText(voiceFile string, vendor string) (text string, err error) {
	start := time.Now()
	defer func() {
		slog.Info(fmt.Sprintf("[VoiceToText] Duration: %dms, voiceFile: 「%s」,text: 「%s」",
			int(time.Since(start).Milliseconds()),
			voiceFile,
			util.EscapeNewline(text),
		))
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

func TextToVoice(text string, voiceFile string, vendor string) (err error) {
	start := time.Now()
	defer func() {
		slog.Info(fmt.Sprintf("[TextToVoice] Duration: %dms, text: 「%s」",
			int(time.Since(start).Milliseconds()),
			text,
		))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
	defer cancel()

	client := getClient(vendor)
	response, err := client.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model: openai.TTSModel1HD,
		Voice: openai.VoiceEcho,
		Input: text,
	})
	if err != nil {
		return
	}

	buf, err := io.ReadAll(response)
	if err != nil {
		return
	}

	err = os.MkdirAll(filepath.Dir(voiceFile), os.ModePerm)
	if err != nil {
		return
	}

	err = os.WriteFile(voiceFile, buf, 0644)
	return
}
