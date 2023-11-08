package handler

import (
	"fmt"
	"net/http"
	"openai/internal/constant"
	"openai/internal/logic"
	"openai/internal/service/errorx"
	"openai/internal/store"
	"strconv"
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
		errorx.RecordError("failed due to invalid msgId", err)
		fmt.Fprint(w, "Invalid msgId")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	exists, _ := store.ReplyChunksExists(msgId)
	if !exists {
		fmt.Fprint(w, "哎呀，消息过期了，重新提问吧~")
		return
	}

	logic.FetchingReply(msgId, func(segment string) {
		fmt.Fprint(w, segment)
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	})
}
