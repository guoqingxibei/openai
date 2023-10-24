package logic

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"openai/internal/constant"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

const oneMonth = 30 * 24 * 3600
const oneWeek = 7 * 24 * 3600

func DecreaseBalance(userName string, mode string) (bool, string) {
	timesPerQuestion := GetTimesPerQuestion(mode)
	if mode == constant.GPT4 || mode == constant.Draw {
		paidBalance, _ := store.GetPaidBalance(userName)
		if paidBalance < timesPerQuestion {
			gpt4BalanceTip := "【余额不足】抱歉，付费次数剩余%d次，不足以继续使用%s模式(每次对话消耗次数%d)，" +
				"可<a href=\"%s\">点我充值次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。" +
				"\n\n%s在此模式下，每次对话仅消耗次数1。"
			return false,
				fmt.Sprintf(gpt4BalanceTip,
					paidBalance,
					mode,
					timesPerQuestion,
					util.GetPayLink(userName),
					util.GetInvitationTutorialLink(),
					getSwitchToGpt3Tip(),
				)
		}

		_, _ = store.DecrPaidBalance(userName, int64(timesPerQuestion))
		return true, ""
	}

	balance := GetBalance(userName)
	if balance < timesPerQuestion {
		paidBalance, _ := store.GetPaidBalance(userName)
		if paidBalance < timesPerQuestion {
			gpt3BalanceTip := "【余额不足】抱歉，你今天的免费次数(%d次)已用完，明天再来吧。费用昂贵，敬请谅解❤️\n\n" +
				"如果使用量很大，可以<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。"
			return false, fmt.Sprintf(gpt3BalanceTip,
				GetQuota(userName),
				util.GetPayLink(userName),
				util.GetInvitationTutorialLink(),
			)
		}
		_, _ = store.DecrPaidBalance(userName, int64(timesPerQuestion))
	} else {
		_, _ = store.DecrBalance(userName, util.Today())
	}
	return true, ""
}

func getSwitchToGpt3Tip() string {
	if util.AccountIsUncle() {
		return "另外，回复「gpt3」可切换到gpt3模式。"
	}
	return "另外，点击菜单「模式-使用gpt3」可切换到gpt3模式。"
}

func calculateQuota(user string) int {
	currentTimestamp := time.Now().Unix()
	subscribeTimestamp, _ := store.GetSubscribeTimestamp(user)
	subscribeInterval := currentTimestamp - subscribeTimestamp
	quota := 0
	if util.AccountIsUncle() {
		if subscribeInterval < oneMonth {
			quota = 10
		} else {
			quota = 1
		}
	} else {
		if subscribeInterval < oneWeek {
			quota = 5
		} else if subscribeInterval < 2*oneWeek {
			quota = 2
		} else {
			quota = 1
		}
	}
	return quota
}

func GetQuota(user string) int {
	today := util.Today()
	quota, err := store.GetQuota(user, today)
	if err == redis.Nil {
		quota = calculateQuota(user)
		_ = store.SetQuota(user, today, quota)
	}
	return quota
}

func GetBalance(user string) int {
	balance, err := fetchBalanceOfToday(user)
	if err != nil {
		if err == redis.Nil {
			quota := GetQuota(user)
			err := SetBalanceOfToday(user, quota)
			if err != nil {
				log.Println("SetBalanceOfToday() failed", err)
				return 0
			}
			return quota
		}
		log.Println("fetchBalanceOfToday() failed", err)
		return 0
	}
	return balance
}

func BuildChatUsage(user string) string {
	return fmt.Sprintf("【次数】免费次数剩余%d次，每天免费%d次。", GetBalance(user), GetQuota(user))
}

func fetchBalanceOfToday(user string) (int, error) {
	return store.GetBalance(user, util.Today())
}

func SetBalanceOfToday(user string, balance int) error {
	return store.SetBalance(user, util.Today(), balance)
}

func AddPaidBalance(user string, added int) int {
	balance, _ := store.GetPaidBalance(user)
	newBalance := added + balance
	_ = store.SetPaidBalance(user, newBalance)
	return newBalance
}
