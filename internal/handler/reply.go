package handler

import (
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"openai/internal/service/gptredis"
)

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

func GetReply(w http.ResponseWriter, r *http.Request) {
	shortMsgId := r.URL.Query().Get("msgId")
	reply, err := gptredis.FetchReply(shortMsgId)
	if err == nil {
		echoJson(w, 0, reply)
	} else if err == redis.Nil {
		echoJson(w, 1, "Not found or expired")
	} else {
		log.Println("GetReply failed", err)
		echoJson(w, 2, "Internal error")
	}
}

func echoJson(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	data, _ := json.Marshal(map[string]interface{}{
		"code":    code,
		"message": message,
	})
	w.Write(data)
}
