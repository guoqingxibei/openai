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
	"openai/internal/service/gptredis"
	"openai/internal/util"
	"strings"
	"time"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type censorResponse struct {
	Conclusion string `json:"conclusion"`
}

var accessToken string
var baiduConfig = config.C.Baidu

func init() {
	token, err := gptredis.FetchBaiduApiAccessToken()
	if err != nil {
		if err != redis.Nil {
			log.Println("gptredis.FetchBaiduApiAccessToken failed", err)
			return
		}
		refreshAccessToken()
	} else {
		if token == "" {
			refreshAccessToken()
		} else {
			accessToken = token
		}
	}

	c := cron.New()
	// Execute once every day at 00:00
	err = c.AddFunc("0 0 0 * * ?", func() {
		refreshAccessToken()
	})
	if err != nil {
		log.Println("AddFunc failed:", err)
		return
	}
	c.Start()
}

func refreshAccessToken() {
	token, expiresIn, err := generateAccessToken()
	if err != nil {
		log.Println("generateAccessToken failed", err)
		return
	}
	log.Println("New Baidu API access token is " + token)
	accessToken = token
	err = gptredis.SetBaiduApiAccessToken(token, time.Second*time.Duration(expiresIn))
	if err != nil {
		log.Println("gptredis.SetBaiduApiAccessToken failed", err)
		return
	}
	log.Println("Refreshed Baidu API access token")
}

func Censor(text string) bool {
	passedChan := make(chan bool, 1)
	go func() {
		passedChan <- censor(text)
	}()

	var passed bool
	select {
	case passed = <-passedChan:
		return passed
	case <-time.After(time.Millisecond * 500):
		log.Printf("[Censor]Skipped the censorship for text:「%s」", util.EscapeNewline(text))
		return true
	}
}

func censor(text string) bool {
	start := time.Now()
	url := "https://aip.baidubce.com/rest/2.0/solution/v1/text_censor/v2/user_defined?access_token=" + accessToken
	payload := strings.NewReader("text=" + text)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		log.Println("http.NewRequest api failed")
		return true
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Timeout: time.Second * 300}
	res, err := client.Do(req)
	if err != nil {
		log.Println("text_censor api failed", err)
		return true
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("ioutil.ReadAll", err)
		return true
	}
	var censorResp censorResponse
	_ = json.Unmarshal(body, &censorResp)
	log.Printf("[CensorAPI]Conclusion: %s, duration: %dms, text:「%s」, detail: %s",
		censorResp.Conclusion,
		int(time.Since(start).Milliseconds()),
		util.EscapeNewline(text),
		string(body),
	)
	return censorResp.Conclusion != "不合规"
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
