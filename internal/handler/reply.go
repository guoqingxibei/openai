package handler

import (
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	replylogic "openai/internal/logic"
	"strconv"
)

func Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

// GetReply Webpage will not retry when code is 1
func GetReply(w http.ResponseWriter, r *http.Request) {
	msgId := r.URL.Query().Get("msgId")
	msgIdInt, err := strconv.ParseInt(msgId, 10, 64)
	if err != nil {
		log.Println("Invalid msgId", err)
		echoJson(w, nil, 1)
		return
	}
	reply, err := replylogic.FetchReply(msgIdInt)
	if err == nil {
		echoJson(w, reply, 0)
	} else if err == redis.Nil {
		// "Not found or expired"
		echoJson(w, nil, 1)
	} else {
		log.Println("GetReply failed", err)
		echoJson(w, nil, -1)
	}
}

func echoJson(w http.ResponseWriter, reply *replylogic.Reply, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(map[string]interface{}{
		"code":  code,
		"reply": reply,
	})
	w.Write(data)
}
