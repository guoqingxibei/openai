package model

type MyError struct {
	ErrorStr           string `json:"err_str"`
	TimestampInSeconds int64  `json:"timestamp_in_seconds"`
}
