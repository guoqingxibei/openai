package wechat

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type MenuResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func initMenu() {
	createOrUpdateMenu()
}

func createOrUpdateMenu() {
	data, err := ioutil.ReadFile("internal/service/wechat/resource/menu.json")
	if err != nil {
		log.Println("failed to open file", err)
		return
	}
	token, err := getAccessToken()
	if err != nil {
		log.Println("failed to get wechat token", err)
		return
	}
	url := "https://api.weixin.qq.com/cgi-bin/menu/create?access_token=" + token
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Println("http.NewRequest failed", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	client := &http.Client{Timeout: time.Second * 300}
	res, err := client.Do(req)
	if err != nil {
		log.Println("create_menu api failed", err)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("ioutil.ReadAll", err)
	}
	var resp MenuResponse
	_ = json.Unmarshal(body, &resp)
	if resp.ErrCode != 0 {
		log.Println("Failed to create menu: " + strconv.Itoa(resp.ErrCode) + " " + resp.ErrMsg)
		return
	}
	log.Println("Refreshed Wechat menu")
}
