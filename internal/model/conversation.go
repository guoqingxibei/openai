package model

import "time"

type Conversation struct {
	Mode     string    `json:"mode"`
	Question string    `json:"question"`
	Answer   string    `json:"answer"`
	Time     time.Time `json:"time"`
}
