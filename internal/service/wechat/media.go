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
	"path/filepath"
	"time"
)

const (
	resourcePath = "internal/service/wechat/resource/images"
)

type uploadResponse struct {
	MediaId string `json:"media_id"`
}

func initMedias() {
	err := filepath.Walk(resourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			initMediaId(filepath.Base(path))
		}

		return nil
	})
	if err != nil {
		log.Println("filepath.Walk error:", err)
	}
}

func initMediaId(imageName string) {
	_, err := gptredis.FetchMediaId(imageName)
	if err != nil {
		if err == redis.Nil {
			_, _ = refreshImage(imageName)
		} else {
			log.Println("gptredis.FetchMediaId failed", err)
		}
	}

	c := cron.New()
	// Execute once every hour
	err = c.AddFunc("0 0 0 * * *", func() {
		_, _ = refreshImage(imageName)
	})
	if err != nil {
		log.Println("AddFunc failed:", err)
		return
	}
	c.Start()
}

func GetMediaId(imageName string) (string, error) {
	mediaId, err := gptredis.FetchMediaId(imageName)
	if err != nil {
		if err == redis.Nil {
			mediaId, err = refreshImage(imageName)
			if err != nil {
				log.Println("refreshImage failed", err)
				return "", err
			}
			return mediaId, nil
		}
		return "", err
	}
	return mediaId, nil
}

func refreshImage(imageName string) (string, error) {
	mediaId, err := uploadImage(imageName)
	if err != nil {
		log.Println("uploadImage failed", err)
		return "", err
	}
	err = gptredis.SetMediaId(mediaId, imageName, time.Hour*24*2)
	if err != nil {
		log.Println("gptredis.SetMediaId failed", err)
		return "", err
	}
	log.Printf("Refreshed the media id of %s", imageName)
	return mediaId, nil
}

func uploadImage(imageName string) (string, error) {
	file, err := os.Open(fmt.Sprintf("%s/%s", resourcePath, imageName))
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
