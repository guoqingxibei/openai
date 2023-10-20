package wechat

import (
	"encoding/json"
	"github.com/silenceper/wechat/v2/officialaccount/menu"
	"io/ioutil"
	"log"
)

type MenuResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func createOrUpdateMenu() {
	data, err := ioutil.ReadFile("internal/service/wechat/resource/menu.json")
	if err != nil {
		log.Println("failed to open file", err)
		return
	}

	var buttons []*menu.Button
	err = json.Unmarshal(data, &buttons)
	if err != nil {
		log.Println("failed to read file", err)
		return
	}

	err = GetAccount().GetMenu().SetMenu(buttons)
	if err != nil {
		log.Println("Failed to refresh menu: ", err)
		return
	}
	log.Println("Refreshed Wechat menu")
}
