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

func CheckBalance(userName string, mode string) (bool, string) {
	if mode == constant.GPT4 {
		paidBalance, _ := store.GetPaidBalance(userName)
		if paidBalance < constant.TimesPerQuestionGPT4 {
			gpt4BalanceTip := "【余额不足】抱歉，付费次数剩余%d次，不足以继续使用gpt4模式(每次提问消耗次数10)，" +
				"可<a href=\"%s\">点我充值次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。" +
				"\n\n%s在此模式下，每次提问仅消耗次数1。"
			return false, fmt.Sprintf(gpt4BalanceTip,
				paidBalance,
				util.GetPayLink(userName),
				util.GetInvitationTutorialLink(),
				getSwitchToGpt3Tip(),
			)
		}
		return true, ""
	}

	balance := GetBalance(userName)
	if balance < 1 {
		paidBalance, _ := store.GetPaidBalance(userName)
		if paidBalance < 1 {
			gpt3BalanceTip := "【余额不足】抱歉，你今天的免费次数(%d次)已用完，明天再来吧。费用昂贵，敬请谅解❤️\n\n" +
				"如果使用量很大，可以<a href=\"%s\">点我购买次数</a>或者<a href=\"%s\">邀请好友获取次数</a>。"
			return false, fmt.Sprintf(gpt3BalanceTip,
				GetQuota(userName),
				util.GetPayLink(userName),
				util.GetInvitationTutorialLink(),
			)
		}
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
				log.Println("store.SetBalance failed", err)
				return 0
			}
			return quota
		}
		log.Println("store.GetBalance failed", err)
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

func DecrBalanceOfToday(user string, mode string) error {
	if mode == constant.GPT4 {
		_, err := store.DecrPaidBalance(user, constant.TimesPerQuestionGPT4)
		return err
	}

	balance := GetBalance(user) // ensure KEY exists before DESC operation while request crosses day
	var err error
	if balance > 0 {
		_, err = store.DecrBalance(user, util.Today())
	} else {
		paidBalance, _ := store.GetPaidBalance(user)
		if paidBalance > 0 {
			_, err = store.DecrPaidBalance(user, 1)
		}
	}
	return err
}

func AddPaidBalance(user string, added int) int {
	balance, _ := store.GetPaidBalance(user)
	newBalance := added + balance
	_ = store.SetPaidBalance(user, newBalance)
	return newBalance
}
