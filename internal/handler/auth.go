package handler

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"openai/internal/config"
	"openai/internal/service/gptredis"
	"time"
)

type tokenResponse struct {
	OpenId string `json:"openid"`
}

func GetOpenId(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	code := r.URL.Query().Get("code")
	openId, _ := gptredis.FetchOpenId(code)
	if openId == "" {
		params := url.Values{}
		params.Add("appid", config.C.Wechat.AppId)
		params.Add("secret", config.C.Wechat.AppSecret)
		params.Add("code", code)
		params.Add("grant_type", "authorization_code")
		fullUrl := "https://api.weixin.qq.com/sns/oauth2/access_token?" + params.Encode()
		res, err := http.Get(fullUrl)
		if err != nil {
			log.Println("oauth2/access_token api failed", err)
			return
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("ioutil.ReadAll", err)
			return
		}
		var tokenResp tokenResponse
		_ = json.Unmarshal(body, &tokenResp)
		log.Printf("[GetTokenAPI] code: %s, openid: %s, duration: %dms, detail: %s",
			code,
			tokenResp.OpenId,
			int(time.Since(start).Milliseconds()),
			string(body),
		)
		openId = tokenResp.OpenId
		_ = gptredis.SetOpenId(code, openId)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(map[string]interface{}{
		"openid": openId,
	})
	w.Write(data)
}
