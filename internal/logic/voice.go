package logic

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/silenceper/wechat/v2/officialaccount/material"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"log"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/service/openaiex"
	"openai/internal/service/wechat"
	"openai/internal/util"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	voiceDir           = "voices"
	maxSegmentDuration = 60
)

func GetTextFromVoice(mediaId string) (text string, err error) {
	mediaURL, err := wechat.GetAccount().GetMaterial().GetMediaURL(mediaId)
	if err != nil {
		return
	}

	voiceFile := genVoiceFileName()
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

func genVoiceFileName() string {
	voiceId := uuid.New().String()
	return fmt.Sprintf("%s/%s.mp3", voiceDir, voiceId)
}

func textToVoice(question string, user string, voiceSentPtr *bool) (err error) {
	voiceFile := genVoiceFileName()
	for _, vendor := range aiVendors {
		err = openaiex.TextToVoice(question, voiceFile, vendor)
		if err == nil {
			break
		}
		log.Printf("openaiex.TextToVoice(%s) failed %v", vendor, err)
	}
	if err != nil {
		return
	}

	duration, err := getAudioDuration(voiceFile)
	if err != nil {
		return
	}

	voiceFiles := []string{voiceFile}
	if duration > maxSegmentDuration {
		voiceFiles, err = splitAudioByDuration(voiceFile, strconv.Itoa(maxSegmentDuration))
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

func getAudioDuration(inputFile string) (duration float64, err error) {
	cmd := exec.Command("ffprobe", "-i", inputFile, "-show_entries", "format=duration", "-v", "quiet", "-of", "csv=p=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}

	return strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
}

func splitAudioByDuration(inputFile string, segmentDuration string) (files []string, err error) {
	outDir := inputFile[:len(inputFile)-4] + "-split-files"
	err = os.Mkdir(outDir, 0755)
	if err != nil {
		return
	}

	cmd := exec.Command("ffmpeg", "-i", inputFile, "-f", "segment", "-segment_time", segmentDuration, outDir+"/%03d.mp3")
	_, err = cmd.CombinedOutput()
	if err != nil {
		return
	}

	err = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return
	}

	sort.Strings(files)
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
