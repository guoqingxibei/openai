package handler

import (
	"encoding/json"
	"net/http"
	"openai/internal/service/recorder"
	"openai/internal/service/wechat"
	"openai/internal/store"
)

func GetOpenId(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	openId, _ := store.GetOpenId(code)
	if openId == "" {
		token, err := wechat.GetAccount().GetOauth().GetUserAccessToken(code)
		if err != nil {
			recorder.RecordError("GetUserAccessToken() failed", err)
			return
		}
		openId = token.OpenID
		_ = store.SetOpenId(code, openId)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(map[string]interface{}{
		"openid": openId,
	})
	w.Write(data)
}
