package logic

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"openai/internal/service/openaiex"
	wechtService "openai/internal/service/wechat"
	"openai/internal/util"
)

const voiceDir = "voices"

func GetTextFromVoice(mediaId string) (text string, err error) {
	mediaURL, err := wechtService.GetAccount().GetMaterial().GetMediaURL(mediaId)
	if err != nil {
		return
	}

	voiceId := uuid.New().String()
	voiceFile := getVoiceFile(voiceId)
	err = util.DownloadFile(mediaURL, voiceFile)
	if err != nil {
		return
	}

	for _, vendor := range aiVendors {
		text, err = openaiex.VoiceToText(voiceFile, vendor)
		if err == nil {
			break
		}
		log.Printf("openaiex.VoiceToText(%s) failed %v", vendor, err)
	}
	return
}

func getVoiceFile(voiceId string) string {
	return fmt.Sprintf("%s/%s.mp3", voiceDir, voiceId)
}
