package model

import "time"

type MyError struct {
	Account string    `json:"account"`
	Title   string    `json:"title"`
	Detail  string    `json:"detail"`
	Time    time.Time `json:"time"`
}
