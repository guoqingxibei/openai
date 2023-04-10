package logic

import (
	"encoding/json"
	"openai/internal/service/gptredis"
)

const (
	Text  = "text"
	Image = "image"
)

type Reply struct {
	ReplyType string `json:"type"`
	Content   string `json:"content"`
	Url       string `json:"url"`
	MediaId   string `json:"mediaId"`
}

func SetEmptyReply(msgId int64) error {
	return gptredis.SetReply(msgId, "")
}

func SetTextReply(msgId int64, content string) error {
	r := Reply{ReplyType: Text, Content: content}
	value, _ := json.Marshal(r)
	return gptredis.SetReply(msgId, string(value))
}

func SetImageReply(msgId int64, url string, mediaId string) error {
	r := Reply{ReplyType: Image, Url: url, MediaId: mediaId}
	value, _ := json.Marshal(r)
	return gptredis.SetReply(msgId, string(value))
}

func FetchReply(msgId int64) (*Reply, error) {
	replyStr, err := gptredis.FetchReply(msgId)
	if err != nil {
		return nil, err
	}
	var r Reply
	_ = json.Unmarshal([]byte(replyStr), &r)
	return &r, nil
}
