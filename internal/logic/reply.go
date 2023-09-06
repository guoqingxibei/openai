package logic

type Reply struct {
	ReplyType string `json:"type"`
	Content   string `json:"content"`
	Url       string `json:"url"`
	MediaId   string `json:"mediaId"`
}
