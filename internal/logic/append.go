package logic

import (
	"log"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
)

const (
	triggerTimes = 50
)

func incUsedTimes(user string) int {
	times, err := gptredis.IncUsedTimes(user)
	if err != nil {
		log.Println("gptredis.IncUsedTimes failed", err)
		return 0
	}
	return times
}

func ShouldAppend(user string) bool {
	times := incUsedTimes(user)
	return times%triggerTimes == 0
}

func selectAppending() string {
	//if rand.Intn(2) == 0 {
	//	return constant.DonateReminder
	//}
	return constant.JoinGroupReminder
}
