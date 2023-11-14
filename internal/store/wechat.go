package store

import (
	"fmt"
	"time"
)

func buildOpenIdKey(authCode string) string {
	return "auth-code:" + authCode + ":open-id"
}

func GetOpenId(authCode string) (string, error) {
	return client.Get(ctx, buildOpenIdKey(authCode)).Result()
}

func SetOpenId(authCode string, openId string) error {
	return client.Set(ctx, buildOpenIdKey(authCode), openId, time.Hour*12).Err()
}
func GetMediaId(imageName string) (string, error) {
	return client.Get(ctx, getMediaIdKey(imageName)).Result()
}

func SetMediaId(mediaId string, mediaName string, expiration time.Duration) error {
	return client.Set(ctx, getMediaIdKey(mediaName), mediaId, expiration).Err()
}

func getMediaIdKey(mediaName string) string {
	return fmt.Sprintf("media-id-of-%s", mediaName)
}
