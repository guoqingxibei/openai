package util

import "time"

func Today() string {
	return time.Now().Format("2006-01-02")
}

func Yesterday() string {
	return time.Now().AddDate(0, 0, -1).Format("2006-01-02")
}

func TimestampToTimeStr(timestampInSeconds int64) string {
	return time.Unix(timestampInSeconds, 0).Format("2006-01-02 15:04:05")
}
