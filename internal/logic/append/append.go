package appendlogic

import (
	"log"
	"openai/internal/service/gptredis"
)

const (
	triggerTimes = 20
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
	return times-1 != 0 && (times-1)%triggerTimes == 0
}

func appendHelpDesc(answer string) string {
	return answer + "\n\n回复 help，可查看使用方法。\n回复 donate，可捐赠作者。"
}

func AppendHelpDescIfPossible(user string, answer string) string {
	if shouldAppendHelpDesc(user) {
		return appendHelpDesc(answer)
	}
	return answer
}
