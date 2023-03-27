package appendlogic

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

func shouldAppendHelpDesc(user string) bool {
	times := incUsedTimes(user)
	return times%triggerTimes == 0
}

func appendHelpDesc(answer string) string {
	return answer + "\n\n" + constant.UsageTail
}

func AppendHelpDescIfPossible(user string, answer string) string {
	if shouldAppendHelpDesc(user) {
		return appendHelpDesc(answer)
	}
	return answer
}
