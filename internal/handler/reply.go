package handler

import (
	"fmt"
	"net/http"
	"openai/internal/constant"
	replylogic "openai/internal/logic"
	"openai/internal/service/errorx"
	"openai/internal/store"
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
		errorx.RecordError("Invalid msgId", err)
		fmt.Fprint(w, "Invalid msgId")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	exists, _ := store.ReplyChunksExists(msgId)
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

		chunks, _ := store.GetReplyChunks(msgId, startIndex, -1)
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
