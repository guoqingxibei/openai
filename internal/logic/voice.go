package logic

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/silenceper/wechat/v2/officialaccount/material"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log/slog"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/service/openaiex"
	"openai/internal/service/wechat"
	"openai/internal/util"
	"strconv"
)

const (
	voiceDir           = constant.Temp + "/voices"
	maxSegmentDuration = 60
)

func GetTextFromVoice(mediaId string) (text string, err error) {
	mediaURL, err := wechat.GetAccount().GetMaterial().GetMediaURL(mediaId)
	if err != nil {
		return
	}

	voiceFile := genVoiceFilePath()
	err = util.DownloadFile(mediaURL, voiceFile)
	if err != nil {
		return
	}

	for _, vendor := range aiVendors {
		text, err = openaiex.VoiceToText(voiceFile, vendor)
		if err == nil {
			break
		}
		slog.Error("openaiex.VoiceToText() failed", "vendor", vendor, "error", err)
	}
	return
}

func genVoiceFilePath() string {
	voiceId := uuid.New().String()
	return fmt.Sprintf("%s/%s.mp3", voiceDir, voiceId)
}

func textToVoice(question string, user string, voiceSentPtr *bool) (err error) {
	voiceFile := genVoiceFilePath()
	for _, vendor := range aiVendors {
		err = openaiex.TextToVoice(question, voiceFile, vendor)
		if err == nil {
			break
		}
		slog.Error("openaiex.TextToVoice() failed", "vendor", vendor, "error", err)
	}
	if err != nil {
		return
	}

	duration, err := util.GetAudioDuration(voiceFile)
	if err != nil {
		return
	}

	voiceFiles := []string{voiceFile}
	if duration > maxSegmentDuration {
		voiceFiles, err = util.SplitAudioByDuration(voiceFile, strconv.Itoa(maxSegmentDuration))
		if err != nil {
			return err
		}
	}

	for _, file := range voiceFiles {
		*voiceSentPtr = true
		err = sendVoiceToUser(file, user)
		if err != nil {
			return err
		}
	}
	return
}

func TextToVoiceEx(question string, user string, voiceSentPtr *bool) (reply string) {
	err := textToVoice(question, user, voiceSentPtr)
	if err != nil {
		AddPaidBalance(user, calTimesForTTS(question))
		errorx.RecordError("textToVoice failed", err)
		reply = constant.TryAgain
	}
	return
}

func sendVoiceToUser(voiceFile string, user string) error {
	media, err := wechat.GetAccount().GetMaterial().MediaUpload(material.MediaTypeVoice, voiceFile)
	if err != nil {
		return err
	}

	return wechat.GetAccount().
		GetCustomerMessageManager().Send(message.NewCustomerVoiceMessage(user, media.MediaID))
}
