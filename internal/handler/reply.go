package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"openai/internal/constant"
	replylogic "openai/internal/logic"
	"openai/internal/service/gptredis"
	"strconv"
	"strings"
	"time"
)

const (
	maxFetchTimes = 6000
)

func GetReplyStream(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprint(w, constant.TryAgain)
		}
	}()

	msgIdStr := r.URL.Query().Get("msgId")
	msgId, err := strconv.ParseInt(msgIdStr, 10, 64)
	if err != nil {
		log.Println("Invalid msgId", err)
		echoJson(w, nil, 1)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	exists, _ := gptredis.ReplyChunksExists(msgId)
	if !exists {
		fmt.Fprint(w, constant.ExpireError)
		return
	}

	var startIndex int64 = 1
	fetchTimes := 0
	for {
		fetchTimes++
		if fetchTimes > maxFetchTimes {
			break
		}

		chunks, _ := gptredis.GetReplyChunks(msgId, startIndex, -1)
		length := len(chunks)
		if length >= 1 {
			reachEnd := chunks[length-1] == replylogic.EndMark
			if reachEnd {
				chunks = chunks[:length-1]
			}
			fmt.Fprint(w, strings.Join(chunks, ""))
			flusher, ok := w.(http.Flusher)
			if ok {
				flusher.Flush()
			}
			startIndex += int64(length)
			if reachEnd {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
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
