package store

import "time"

func GetBaiduApiAccessToken() (string, error) {
	return client.Get(ctx, getBaiduApiAccessTokenKey()).Result()
}

func SetBaiduApiAccessToken(accessToken string, expiration time.Duration) error {
	return client.Set(ctx, getBaiduApiAccessTokenKey(), accessToken, expiration).Err()
}

func getBaiduApiAccessTokenKey() string {
	return "baidu-api-access-token"
}
