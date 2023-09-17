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

func CheckBalance(inMsg *wechat.Msg, gptMode string) (bool, string) {
	userName := inMsg.FromUserName
	if gptMode == constant.GPT4 {
		paidBalance, _ := gptredis.FetchPaidBalance(userName)
		if paidBalance < constant.TimesPerQuestionGPT4 {
			gpt4BalanceTip := "【余额不足】抱歉，付费次数剩余%d次，不足以继续使用gpt4模式(每次提问消耗次数10)，" +
				"可<a href=\"https://brother.cxyds.top/shop?uncle_openid=%s\">点我充值次数</a>。" +
				"\n\n另外，回复gpt3可切换到gpt3模式。在此模式下，每次提问仅消耗次数1。"
			return false, fmt.Sprintf(gpt4BalanceTip, paidBalance, userName)
		}
		return true, ""
	}

	balance := GetBalance(userName)
	if balance < 1 {
		paidBalance, _ := gptredis.FetchPaidBalance(userName)
		if paidBalance < 1 {
			gpt3BalanceTip := "【余额不足】抱歉，你今天的免费次数(%d次)已用完，明天再来吧。费用昂贵，敬请谅解❤️\n\n" +
				"如果使用量很大，可以<a href=\"https://brother.cxyds.top/shop?uncle_openid=%s\">点我购买次数</a>。"
			return false, fmt.Sprintf(gpt3BalanceTip, GetQuota(userName), userName)
		}
	}
	return true, ""
}

func calculateQuota(user string) int {
	currentTimestamp := time.Now().Unix()
	subscribeTimestamp, _ := gptredis.FetchSubscribeTimestamp(user)
	subscribeInterval := currentTimestamp - subscribeTimestamp
	quota := 0
	if subscribeInterval < oneMonth {
		quota = 10
	} else if subscribeInterval < 2*oneMonth {
		quota = 1
	} else {
		quota = 1
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

func GetBalance(user string) int {
	balance, err := fetchBalanceOfToday(user)
	if err != nil {
		if err == redis.Nil {
			quota := GetQuota(user)
			err := setBalanceOfToday(user, quota)
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
	return fmt.Sprintf("【次数】免费次数剩余%d次，每天免费%d次。", GetBalance(user), GetQuota(user))
}

func fetchBalanceOfToday(user string) (int, error) {
	return gptredis.FetchBalance(user, util.Today())
}

func setBalanceOfToday(user string, balance int) error {
	return gptredis.SetBalance(user, util.Today(), balance)
}

func DecrBalanceOfToday(user string, gptMode string) error {
	if gptMode == constant.GPT4 {
		_, err := gptredis.DecrPaidBalance(user, constant.TimesPerQuestionGPT4)
		return err
	}

	balance := GetBalance(user) // ensure KEY exists before DESC operation while request crosses day
	var err error
	if balance > 0 {
		_, err = gptredis.DecrBalance(user, util.Today())
	} else {
		paidBalance, _ := gptredis.FetchPaidBalance(user)
		if paidBalance > 0 {
			_, err = gptredis.DecrPaidBalance(user, 1)
		}
	}
	return err
}
