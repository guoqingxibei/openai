package logic

import (
	"log"
	"openai/internal/constant"
	"openai/internal/store"
)

const (
	triggerTimes = 20
)

func incUsedTimes(user string) int {
	times, err := store.IncUsedTimes(user)
	if err != nil {
		log.Println("store.IncUsedTimes() failed", err)
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
