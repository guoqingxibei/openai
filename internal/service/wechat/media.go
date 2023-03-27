package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"openai/internal/service/gptredis"
	"os"
	"time"
)

type uploadResponse struct {
	MediaId string `json:"media_id"`
}

func initDonateQrMediaId() {
	_, err := gptredis.FetchMediaIdOfDonateQr()
	if err != nil {
		if err == redis.Nil {
			_, _ = refreshDonateQrImage()
		} else {
			log.Println("gptredis.FetchMediaIdOfDonateQr failed", err)
		}
	}

	c := cron.New()
	// Execute once every hour
	err = c.AddFunc("0 0 0 * * *", func() {
		_, _ = refreshDonateQrImage()
	})
	if err != nil {
		log.Println("AddFunc failed:", err)
		return
	}
	c.Start()
}

func GetMediaIdOfDonateQr() (string, error) {
	mediaId, err := gptredis.FetchMediaIdOfDonateQr()
	if err != nil {
		if err == redis.Nil {
			mediaId, err = refreshDonateQrImage()
			if err != nil {
				log.Println("refreshDonateQrImage failed", err)
				return "", err
			}
			return mediaId, nil
		}
		return "", err
	}
	return mediaId, nil
}

func refreshDonateQrImage() (string, error) {
	mediaId, err := uploadDonateQrImage()
	if err != nil {
		log.Println("uploadDonateQrImage failed", err)
		return "", err
	}
	err = gptredis.SetMediaIdOfDonateQr(mediaId, time.Hour*24*2)
	if err != nil {
		log.Println("gptredis.SetMediaIdOfDonateQr failed", err)
		return "", err
	}
	log.Println("Refreshed the media id of donate qr")
	return mediaId, nil
}

func uploadDonateQrImage() (string, error) {
	file, err := os.Open("internal/service/wechat/resource/donate_qr.JPG")
	if err != nil {
		log.Println("failed to open file", err)
		return "", err
	}
	defer file.Close()

	return uploadToWechat(file)
}

func UploadImageFromUrl(url string) (string, error) {
	start := time.Now()
	imageResp, err := http.Get(url)
	log.Printf("[GetImageAPI] Duration: %dms, image url: %s",
		int(time.Since(start).Milliseconds()),
		url,
	)
	if err != nil {
		return "", err
	}
	defer imageResp.Body.Close()

	return uploadToWechat(imageResp.Body)
}

func uploadToWechat(src io.Reader) (string, error) {
	start := time.Now()
	token, err := getAccessToken()
	if err != nil {
		return "", err
	}

	uploadUrl := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s",
		token,
		"image",
	)

	uploadBody := &bytes.Buffer{}
	writer := multipart.NewWriter(uploadBody)
	imageFileName := uuid.NewString() + ".jpg"
	part, err := writer.CreateFormFile("image", imageFileName)
	if err != nil {
		fmt.Println("Error creating form file:", err)
		return "", err
	}
	io.Copy(part, src)
	writer.Close()

	req, err := http.NewRequest("post", uploadUrl, uploadBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{Timeout: time.Second * 300}
	uploadRes, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer uploadRes.Body.Close()

	body, err := ioutil.ReadAll(uploadRes.Body)
	if err != nil {
		return "", err
	}
	var uploadResp uploadResponse
	_ = json.Unmarshal(body, &uploadResp)
	mediaId := uploadResp.MediaId
	log.Printf("[UploadImageAPI] Duration: %dms, media id: %s, image name: %s",
		int(time.Since(start).Milliseconds()),
		mediaId,
		imageFileName,
	)
	return mediaId, nil
}
