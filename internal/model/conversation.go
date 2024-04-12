package model

import "time"

type Conversation struct {
	Mode        string    `json:"mode"`
	PaidBalance int       `json:"paid_balance"`
	Question    string    `json:"question"`
	Answer      string    `json:"answer"`
	Time        time.Time `json:"time"`
}
