package logic

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"openai/internal/constant"
	"openai/internal/service/gptredis"
	"openai/internal/service/wechat"
	"openai/internal/util"
	"time"
)

const oneMonth = 30 * 24 * 3600

func CheckBalance(inMsg *wechat.Msg, mode string) bool {
	userName := inMsg.FromUserName
	balance := GetBalance(userName, mode)
	if balance <= 0 {
		paidBalance, _ := gptredis.FetchPaidBalance(userName)
		if paidBalance <= 0 {
			return false
		}
	}

	return true
}

func calculateQuota(user string) int {
	return 5

	currentTimestamp := time.Now().Unix()
	subscribeTimestamp, _ := gptredis.FetchSubscribeTimestamp(user)
	subscribeInterval := currentTimestamp - subscribeTimestamp
	quota := 0
	if subscribeInterval < oneMonth {
		quota = 10
	} else if subscribeInterval < 6*oneMonth {
		quota = 5
	} else {
		quota = 2
	}
	return quota
}

func GetQuota(user string) int {
	today := util.Today()
	quota, err := gptredis.GetQuota(user, today)
	if err == redis.Nil {
		quota = calculateQuota(user)
		_ = gptredis.SetQuota(user, today, quota)
	}
	return quota
}

func GetBalance(user string, mode string) int {
	balance, err := fetchBalanceOfToday(user, mode)
	if err != nil {
		if err == redis.Nil {
			quota := GetQuota(user)
			err := setBalanceOfToday(user, mode, quota)
			if err != nil {
				log.Println("gptredis.SetBalance failed", err)
				return 0
			}
			return quota
		}
		log.Println("gptredis.GetBalance failed", err)
		return 0
	}
	return balance
}

func BuildChatUsage(user string) string {
	return fmt.Sprintf(constant.ChatUsage, GetQuota(user), GetBalance(user, constant.Chat))
}

func fetchBalanceOfToday(user string, mode string) (int, error) {
	return gptredis.FetchBalance(user, mode, util.Today())
}

func setBalanceOfToday(user string, mode string, balance int) error {
	return gptredis.SetBalance(user, mode, util.Today(), balance)
}

func DecrBalanceOfToday(user string, mode string) error {
	balance := GetBalance(user, mode) // ensure KEY exists before DESC operation while request crosses day
	var err error
	if balance > 0 {
		_, err = gptredis.DecrBalance(user, mode, util.Today())
	} else {
		paidBalance, _ := gptredis.FetchPaidBalance(user)
		if paidBalance > 0 {
			_, err = gptredis.DecrPaidBalance(user)
		}
	}
	return err
}
