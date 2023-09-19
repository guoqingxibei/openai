package sb

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"openai/internal/config"
	"openai/internal/util"
	"strconv"
	"time"
)

type statusResp struct {
	Data struct {
		Credit string `json:"credit"`
	} `json:"data"`
}

func GetSbBalance() (float64, error) {
	start := time.Now()
	params := url.Values{}
	params.Add("api_key", config.C.OpenaiSb.Key)
	fullUrl := "https://api.openai-sb.com/sb-api/user/status?" + params.Encode()
	res, err := http.Get(fullUrl)
	if err != nil {
		return 0, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	var resp statusResp
	_ = json.Unmarshal(body, &resp)
	log.Printf("[GetSbStatusAPI] Duration: %dms, response: %s",
		int(time.Since(start).Milliseconds()),
		util.EscapeNewline(string(body)),
	)
	credit, err := strconv.ParseFloat(resp.Data.Credit, 32)
	return credit / 10000, nil
}
