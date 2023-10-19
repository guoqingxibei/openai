package ohmygpt

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"openai/internal/config"
	"openai/internal/util"
	"strconv"
	"strings"
	"time"
)

type balanceResponse struct {
	Data struct {
		Balance string `json:"balance"`
	} `json:"data"`
}

func GetOhmygptBalance() (float64, error) {
	start := time.Now()
	url := config.C.Ohmygpt.BaseURL + "/api/v1/user/admin/balance"
	payload := strings.NewReader("")
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return 0, err
	}
	req.Header.Add("Authorization", "Bearer "+config.C.Ohmygpt.Key)
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
	var resp balanceResponse
	_ = json.Unmarshal(body, &resp)
	log.Printf("[GetOhmygptBalanceAPI] Duration: %dms, response: 「%s」",
		int(time.Since(start).Milliseconds()),
		util.EscapeNewline(string(body)),
	)
	balanceFloat, err := strconv.ParseFloat(resp.Data.Balance, 32)
	return balanceFloat / 34000, nil
}
