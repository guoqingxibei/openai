package wechat

import (
	"encoding/json"
	"github.com/silenceper/wechat/v2/officialaccount/menu"
	"io/ioutil"
	"log"
	"openai/internal/service/recorder"
)

type MenuResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func createOrUpdateMenu() {
	data, err := ioutil.ReadFile("internal/service/wechat/resource/menu.json")
	if err != nil {
		recorder.RecordError("ioutil.ReadFile() failed", err)
		return
	}

	var buttons []*menu.Button
	err = json.Unmarshal(data, &buttons)
	if err != nil {
		recorder.RecordError("json.Unmarshal() failed", err)
		return
	}

	err = GetAccount().GetMenu().SetMenu(buttons)
	if err != nil {
		recorder.RecordError("SetMenu() failed", err)
		return
	}
	log.Println("Refreshed wechat menu")
}
