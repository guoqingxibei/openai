package logic

import (
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"openai/internal/constant"
	"openai/internal/service/errorx"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

const oneMonth = 30 * 24 * 3600
const oneWeek = 7 * 24 * 3600

func DecreaseBalance(userName string, mode string, question string) (bool, string) {
	timesPerQuestion := GetTimesPerQuestion(mode)
	paidBalance, err := store.GetPaidBalance(userName)
	hasPaid := !errors.Is(err, redis.Nil)

	if mode == constant.TTS {
		times := calTimesForTTS(question)
		if paidBalance < times {
			balanceTip := "【余额不足】抱歉，付费额度剩余%d次，而此次语音转换需要消耗%d次，" +
				"请<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。"
			return false, fmt.Sprintf(balanceTip,
				paidBalance,
				times,
				util.GetPayLink(userName),
				util.GetInvitationTutorialLink(),
			)
		}

		_, _ = store.DecrPaidBalance(userName, int64(times))
		return true, ""
	}

	if mode == constant.Draw {
		if paidBalance < timesPerQuestion {
			drawBalanceTip := "【余额不足】抱歉，付费额度剩余%d次，不足以继续使用%s模式(每次绘画消耗次数%d)，" +
				"<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。" +
				"\n\n%s在此模式下，每次对话仅消耗次数1。"
			return false,
				fmt.Sprintf(drawBalanceTip,
					paidBalance,
					GetModeName(mode),
					timesPerQuestion,
					util.GetPayLink(userName),
					util.GetInvitationTutorialLink(),
					getSwitchToGpt3Tip(),
				)
		}

		_, _ = store.DecrPaidBalance(userName, int64(timesPerQuestion))
		return true, ""
	}

	if mode == constant.GPT4 || mode == constant.GPT4Dot5 {
		if paidBalance < timesPerQuestion {
			gpt4BalanceTip := "【余额不足】抱歉，付费额度剩余%d次，不足以继续使用%s模式(每次对话消耗次数%d)，" +
				"<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。" +
				"\n\n%s在此模式下，每次对话仅消耗次数1。"
			return false,
				fmt.Sprintf(gpt4BalanceTip,
					paidBalance,
					GetModeName(mode),
					timesPerQuestion,
					util.GetPayLink(userName),
					util.GetInvitationTutorialLink(),
					getSwitchToGpt3Tip(),
				)
		}

		_, _ = store.DecrPaidBalance(userName, int64(timesPerQuestion))
		return true, ""
	}

	if mode == constant.DeepSeekR1 {
		if paidBalance < timesPerQuestion {
			deepSeekR1BalanceTip := "【余额不足】抱歉，付费额度剩余%d次，不足以继续使用%s模式(每次对话消耗次数%d)，" +
				"<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。" +
				"\n\n%s在此模式下，每次对话仅消耗次数1。"
			return false,
				fmt.Sprintf(deepSeekR1BalanceTip,
					paidBalance,
					GetModeName(mode),
					timesPerQuestion,
					util.GetPayLink(userName),
					util.GetInvitationTutorialLink(),
					getSwitchToGpt3Tip(),
				)
		}

		_, _ = store.DecrPaidBalance(userName, int64(timesPerQuestion))
		return true, ""
	}

	if mode == constant.GPT3 {
		if hasPaid {
			if paidBalance < timesPerQuestion {
				gpt3BalanceTip := "【余额不足】抱歉，你的付费额度已用完，" +
					"请<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数。"
				return false, fmt.Sprintf(gpt3BalanceTip,
					util.GetPayLink(userName),
					util.GetInvitationTutorialLink(),
				)
			}
			_, _ = store.DecrPaidBalance(userName, int64(timesPerQuestion))
		} else {
			balance := GetBalance(userName)
			if balance < timesPerQuestion {
				gpt3BalanceTip := "【余额不足】抱歉，你今天的免费额度(%d次)已用完，明天再来吧。费用昂贵，敬请谅解❤️\n\n" +
					"如果使用量大，可以<a href=\"%s\">点我购买</a>或者<a href=\"%s\">邀请好友</a>获取次数，次数永久有效。"
				return false, fmt.Sprintf(gpt3BalanceTip,
					GetQuota(userName),
					util.GetPayLink(userName),
					util.GetInvitationTutorialLink(),
				)
			}
			_, _ = store.DecrBalance(userName, util.Today())
		}
		return true, ""
	}

	// default
	errorx.RecordError("[DecreaseBalance] failed with unknown mode", errors.New("unknown mode: "+mode))
	return true, ""
}

func getSwitchToGpt3Tip() string {
	if util.AccountIsUncle() {
		return "温馨提醒，回复「gpt3」可切换到gpt3模式。"
	}
	return "温馨提醒，点击菜单「模式-使用gpt3」可切换到gpt3模式。"
}

func calculateQuota(user string) int {
	currentTimestamp := time.Now().Unix()
	subscribeTimestamp, _ := store.GetSubscribeTimestamp(user)
	subscribeInterval := currentTimestamp - subscribeTimestamp
	quota := 0
	if subscribeInterval < oneWeek {
		quota = 5
	} else {
		quota = 2
	}
	return quota
}

func GetQuota(user string) int {
	today := util.Today()
	quota, err := store.GetQuota(user, today)
	if errors.Is(err, redis.Nil) {
		quota = calculateQuota(user)
		_ = store.SetQuota(user, today, quota)
	}
	return quota
}

func GetBalance(user string) int {
	balance, err := fetchBalanceOfToday(user)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			quota := GetQuota(user)
			err := SetBalanceOfToday(user, quota)
			if err != nil {
				slog.Error("SetBalanceOfToday() failed", "error", err)
				return 0
			}
			return quota
		}
		slog.Error("fetchBalanceOfToday() failed", "error", err)
		return 0
	}
	return balance
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
