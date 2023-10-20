package wechat

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
	"github.com/silenceper/wechat/v2/officialaccount/material"
	"log"
	"openai/internal/store"
	"os"
	"path/filepath"
	"time"
)

const (
	resourcePath = "internal/service/wechat/resource/images"
)

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
	_, err := store.GetMediaId(imageName)
	if err != nil {
		if err == redis.Nil {
			_, _ = refreshImage(imageName)
		} else {
			log.Println("store.GetMediaId failed", err)
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
	mediaId, err := store.GetMediaId(imageName)
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
	err = store.SetMediaId(mediaId, imageName, time.Hour*24*2)
	if err != nil {
		log.Println("store.SetMediaId failed", err)
		return "", err
	}
	log.Printf("Refreshed the media id of %s", imageName)
	return mediaId, nil
}

func uploadImage(imageName string) (string, error) {
	fileName := fmt.Sprintf("%s/%s", resourcePath, imageName)
	media, err := GetAccount().GetMaterial().MediaUpload(material.MediaTypeImage, fileName)
	if err != nil {
		log.Println("MediaUpload failed", err)
		return "", err
	}
	return media.MediaID, nil
}
