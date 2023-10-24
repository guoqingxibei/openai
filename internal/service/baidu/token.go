package baidu

import (
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"net/http"
	"openai/internal/config"
	"openai/internal/service/errorx"
	"openai/internal/store"
	"strings"
	"time"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

var baiduConfig = config.C.Baidu

func init() {
	_, err := store.GetBaiduApiAccessToken()
	if err != nil {
		if err == redis.Nil {
			_, err := refreshAccessToken()
			if err != nil {
				errorx.RecordError("refreshAccessToken() failed", err)
			}
		} else {
			errorx.RecordError("store.GetBaiduApiAccessToken() failed", err)
		}
	}

	c := cron.New()
	// Execute once every day at 00:00
	err = c.AddFunc("0 0 0 * * ?", func() {
		_, err := refreshAccessToken()
		if err != nil {
			errorx.RecordError("refreshAccessToken() failed", err)
		}
	})
	if err != nil {
		errorx.RecordError("AddFunc() failed", err)
		return
	}
	c.Start()
}

func refreshAccessToken() (string, error) {
	token, expiresIn, err := generateAccessToken()
	if err != nil {
		errorx.RecordError("generateAccessToken() failed", err)
		return "", err
	}
	log.Println("New Baidu API access token is " + token)
	err = store.SetBaiduApiAccessToken(token, time.Second*time.Duration(expiresIn))
	if err != nil {
		log.Println("store.SetBaiduApiAccessToken failed", err)
		return "", err
	}
	log.Println("Refreshed Baidu API access token")
	return token, nil
}

func generateAccessToken() (string, int, error) {
	url := "https://aip.baidubce.com/oauth/2.0/token"
	postData := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		baiduConfig.ApiKey, baiduConfig.SecretKey)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(postData))
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
	return tokenResp.AccessToken, tokenResp.ExpiresIn, nil
}

func getAccessToken() (string, error) {
	token, err := store.GetBaiduApiAccessToken()
	if err != nil {
		if err == redis.Nil {
			return refreshAccessToken()
		}
		return "", err
	}
	return token, nil
}
