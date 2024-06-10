package wechat

import (
	"encoding/json"
	"github.com/silenceper/wechat/v2/officialaccount/menu"
	"io/ioutil"
	"log"
	"openai/internal/service/errorx"
)

type MenuResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func createOrUpdateMenu() {
	data, err := ioutil.ReadFile("resource/menu.json")
	if err != nil {
		errorx.RecordError("ioutil.ReadFile() failed", err)
		return
	}

	var buttons []*menu.Button
	err = json.Unmarshal(data, &buttons)
	if err != nil {
		errorx.RecordError("json.Unmarshal() failed", err)
		return
	}

	err = GetAccount().GetMenu().SetMenu(buttons)
	if err != nil {
		errorx.RecordError("SetMenu() failed", err)
		return
	}
	log.Println("Refreshed wechat menu")
}
