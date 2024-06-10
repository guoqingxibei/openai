package baidu

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"openai/internal/service/errorx"
	"openai/internal/util"
	"strings"
	"time"
)

type censorResponse struct {
	Conclusion string `json:"conclusion"`
}

func Censor(text string) bool {
	if text == "" {
		return true
	}

	start := time.Now()
	token, err := getAccessToken()
	if err != nil {
		return true
	}
	url := "https://aip.baidubce.com/rest/2.0/solution/v1/text_censor/v2/user_defined?access_token=" + token
	payload := strings.NewReader("text=" + text)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		errorx.RecordError("http.NewRequest() failed", err)
		return true
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Timeout: time.Second * 300}
	res, err := client.Do(req)
	if err != nil {
		errorx.RecordError("text_censor API failed", err)
		return true
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		errorx.RecordError("ioutil.ReadAll() failed", err)
		return true
	}
	var censorResp censorResponse
	_ = json.Unmarshal(body, &censorResp)
	slog.Info(fmt.Sprintf("[CensorAPI] Conclusion: %s, duration: %dms, text: 「%s」, detail: %s",
		censorResp.Conclusion,
		int(time.Since(start).Milliseconds()),
		util.EscapeNewline(text),
		string(body),
	))
	return censorResp.Conclusion == "" || censorResp.Conclusion == "合规"
}
