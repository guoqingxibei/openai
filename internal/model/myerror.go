package model

import "time"

type MyError struct {
	Title  string    `json:"title"`
	Detail string    `json:"detail"`
	Time   time.Time `json:"time"`
}
