package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"openai/internal/service/wechat"
	"openai/internal/store"
)

func GetOpenId(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	openId, _ := store.GetOpenId(code)
	if openId == "" {
		token, err := wechat.GetAccount().GetOauth().GetUserAccessToken(code)
		if err != nil {
			log.Println("GetUserAccessToken() failed", err)
			return
		}
		_ = store.SetOpenId(code, token.OpenID)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(map[string]interface{}{
		"openid": openId,
	})
	w.Write(data)
}
