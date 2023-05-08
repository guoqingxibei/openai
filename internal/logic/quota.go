package logic

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"time"
)

const boundaryTimestamp = 1681315200 // Thu Apr 13 2023 00:00:00 GMT+0800 (China Standard Time)
const oldChatQuota = 40

var quotaMap = map[string]int{
	constant.Chat:  40,
	constant.Image: 1,
}

func getChatQuota(user string) int {
	timestamp, _ := gptredis.FetchSubscribeTimestamp(user)
	if timestamp < boundaryTimestamp {
		return oldChatQuota
	}
	return quotaMap[constant.Chat]
}

func CheckBalance(inMsg *wechat.Msg, mode string) (bool, string) {
	balance := FetchBalance(inMsg.FromUserName, mode)
	if balance <= 0 {
		msg := constant.ZeroChatBalance
		if mode == constant.Image {
			msg = constant.ZeroImageBalance
		}
		return false, msg
	}

	return true, ""
}

func FetchBalance(user string, mode string) int {
	quota := quotaMap[mode]
	balance, err := fetchBalanceOfToday(user, mode)
	if err != nil {
		if err == redis.Nil {
			if mode == constant.Chat {
				quota = getChatQuota(user)
			}
			err := setBalanceOfToday(user, mode, quota)
			if err != nil {
				log.Println("gptredis.SetBalance failed", err)
				return 0
			}
			return quota
		}
		log.Println("gptredis.FetchBalance failed", err)
		return 0
	}
	return balance
}

func BuildImageUsage(user string) string {
	return fmt.Sprintf(constant.ImageUsage, quotaMap[constant.Image], FetchBalance(user, constant.Image))
}

func BuildChatUsage(user string) string {
	return fmt.Sprintf(constant.ChatUsage, getChatQuota(user), FetchBalance(user, constant.Chat))
}

func fetchBalanceOfToday(user string, mode string) (int, error) {
	return gptredis.FetchBalance(user, mode, today())
}

func setBalanceOfToday(user string, mode string, balance int) error {
	return gptredis.SetBalance(user, mode, today(), balance)
}

func DecrBalanceOfToday(user string, mode string) (int, error) {
	FetchBalance(user, mode) // ensure KEY exists before DESC operation while request crosses day
	return gptredis.DecrBalance(user, mode, today())
}

func today() string {
	return time.Now().Format("2006-01-02")
}
