package baidu

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"openai/internal/util"
	"strings"
	"time"
)

type censorResponse struct {
	Conclusion string `json:"conclusion"`
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
		log.Printf("[Censor] Skipped the censorship for text:「%s」", util.EscapeNewline(text))
		return true
	}
}

func censor(text string) bool {
	start := time.Now()
	token, err := getAccessToken()
	if err != nil {
		return true
	}
	url := "https://aip.baidubce.com/rest/2.0/solution/v1/text_censor/v2/user_defined?access_token=" + token
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
	log.Printf("[CensorAPI] Conclusion: %s, duration: %dms, text:「%s」, detail: %s",
		censorResp.Conclusion,
		int(time.Since(start).Milliseconds()),
		util.EscapeNewline(text),
		string(body),
	)
	return censorResp.Conclusion != "不合规"
}
