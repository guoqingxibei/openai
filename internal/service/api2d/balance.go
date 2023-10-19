package api2d

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"openai/internal/config"
	"openai/internal/util"
	"strings"
	"time"
)

type profileResponse struct {
	Data struct {
		Profile struct {
			Point int `json:"point"`
		} `json:"profile"`
	} `json:"data"`
}

func GetApi2dBalance() (float64, error) {
	start := time.Now()
	url := "https://api.api2win.com/user/profile"
	payload := strings.NewReader("")
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Authorization", "Bearer "+config.C.OpenaiApi2d.Token)
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Timeout: time.Second * 300}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	var resp profileResponse
	_ = json.Unmarshal(body, &resp)
	log.Printf("[GetApi2dPointAPI] Duration: %dms, response: 「%s」",
		int(time.Since(start).Milliseconds()),
		util.EscapeNewline(string(body)),
	)
	point := resp.Data.Profile.Point
	return float64(point) * 0.0021, nil
}
