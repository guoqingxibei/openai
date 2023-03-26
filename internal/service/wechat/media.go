package wechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"time"
)

type uploadResponse struct {
	MediaId string `json:"media_id"`
}

func UploadImage(url string) (string, error) {
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

	token, err := getAccessToken()
	if err != nil {
		return "", err
	}

	start = time.Now()
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
	io.Copy(part, imageResp.Body)
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
	log.Printf("[UploadImageAPI] Duration: %dms, image url: %s, media id: %s, image name: %s",
		int(time.Since(start).Milliseconds()),
		url,
		mediaId,
		imageFileName,
	)
	return mediaId, nil
}
