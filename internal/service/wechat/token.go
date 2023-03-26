package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"net/http"
	"openai/internal/config"
	"openai/internal/service/gptredis"
	"time"
)

var wechatConfig = config.C.Wechat

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func init() {
	_, err := gptredis.FetchWechatApiAccessToken()
	if err != nil {
		if err == redis.Nil {
			_, err := refreshAccessToken()
			if err != nil {
				log.Println("refreshAccessToken failed", err)
			}
		} else {
			log.Println("gptredis.FetchWechatApiAccessToken failed", err)
		}
	}

	c := cron.New()
	// Execute once every hour
	err = c.AddFunc("0 0 * * * *", func() {
		_, err := refreshAccessToken()
		if err != nil {
			log.Println("refreshAccessToken failed", err)
		}
	})
	if err != nil {
		log.Println("AddFunc failed:", err)
		return
	}
	c.Start()
}

func refreshAccessToken() (string, error) {
	token, expiresIn, err := generateAccessToken()
	if err != nil {
		log.Println("generateAccessToken failed", err)
		return "", err
	}
	log.Println("New Wechat API access token is " + token)
	err = gptredis.SetWechatApiAccessToken(token, time.Second*time.Duration(expiresIn))
	if err != nil {
		log.Println("gptredis.SetWechatApiAccessToken failed", err)
		return "", err
	}
	log.Println("Refreshed Wechat API access token")
	return token, nil
}

func generateAccessToken() (string, int, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		wechatConfig.AppId,
		wechatConfig.AppSecret,
	)
	resp, err := http.Get(url)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	var tokenResp tokenResponse
	_ = json.Unmarshal(body, &tokenResp)
	if tokenResp.AccessToken == "" {
		return "", 0, errors.New(string(body))
	}
	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}

func getAccessToken() (string, error) {
	token, err := gptredis.FetchWechatApiAccessToken()
	if err != nil {
		if err == redis.Nil {
			return refreshAccessToken()
		}
		return "", err
	}
	return token, nil
}
