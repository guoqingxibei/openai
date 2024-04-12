package logic

import (
	"fmt"
	"github.com/robfig/cron"
	"log"
	"openai/internal/constant"
	"openai/internal/service/api2d"
	"openai/internal/service/email"
	"openai/internal/service/errorx"
	"openai/internal/service/ohmygpt"
	"openai/internal/service/sb"
	"openai/internal/store"
	"openai/internal/util"
)

func init() {
	c1 := cron.New()
	// Execute once every day at 00:00
	err := c1.AddFunc("0 0 0 * * ?", func() {
		sendYesterdayReportEmail()
	})
	if err != nil {
		errorx.RecordError("AddFunc() failed", err)
		return
	}
	c1.Start()

	if util.AccountIsBrother() && util.EnvIsProd() {
		c2 := cron.New()
		// Execute once every hour
		err = c2.AddFunc("0 0 * * * *", func() {
			checkVendorBalance()
		})
		if err != nil {
			errorx.RecordError("AddFunc() failed", err)
			return
		}
		c2.Start()
	}
}

func checkVendorBalance() {
	log.Println("Checking balance of vendors...")
	alarm := false
	ohmygptBalance, _ := ohmygpt.GetOhmygptBalance()
	if ohmygptBalance < 30 {
		alarm = true
	}
	sbBalance, _ := sb.GetSbBalance()
	if sbBalance < 0.1 {
		alarm = true
	}
	api2dBalance, _ := api2d.GetApi2dBalance()
	if api2dBalance < 1 {
		alarm = true
	}
	if alarm {
		log.Println("Balance is insufficient, sending email...")
		body := fmt.Sprintf("Ohmygpt: ￥%.2f\nSB: ￥%.2f\nApi2d: ￥%.2f",
			ohmygptBalance, sbBalance, api2dBalance)
		email.SendEmail("Insufficient Balance", body)
	}
	log.Println("Check finished")
}

func sendYesterdayReportEmail() {
	yesterday := util.Yesterday()
	subject := fmt.Sprintf("[%s/%s] Summary for %s", util.GetAccount(), util.GetEnv(), yesterday)

	body := ""
	ohmygptBalance, _ := ohmygpt.GetOhmygptBalance()
	sbBalance, _ := sb.GetSbBalance()
	api2dBalance, _ := api2d.GetApi2dBalance()
	balanceTitle := "[balance]\n"
	balanceContent := fmt.Sprintf("Ohmygpt: ￥%.2f\nSB: ￥%.2f\nApi2d: ￥%.2f\n",
		ohmygptBalance, sbBalance, api2dBalance)
	body += balanceTitle + balanceContent

	tradeNos, _ := store.GetSuccessOutTradeNos(yesterday)
	transactionTitle := fmt.Sprintf("\n[%d transactions]\n", len(tradeNos))
	transactionBody := ""
	for _, tradeNo := range tradeNos {
		transaction, _ := store.GetTransaction(tradeNo)
		paidAccount := constant.Brother
		if transaction.UncleOpenId != "" {
			paidAccount = constant.Uncle
		}
		transactionLine := fmt.Sprintf(
			"%s ￥%d %s\n",
			util.FormatTime(transaction.UpdatedAt),
			transaction.PriceInFen/100,
			paidAccount,
		)
		transactionBody = transactionLine + transactionBody
	}
	body += transactionTitle + transactionBody

	errCnt, errorContent := errorx.GetErrorsDesc(yesterday)
	errorTitle := fmt.Sprintf("\n[%d errors]\n", errCnt)
	body += errorTitle + errorContent

	users, _ := store.GetActiveUsers(yesterday)
	userCnt := len(users)
	convCnt := 0
	convContent := ""
	for idx, user := range users {
		convContent += fmt.Sprintf("==%d/%d %s==\n", idx+1, userCnt, user)
		convs, _ := store.GetConversations(user, yesterday)
		for convIdx, conv := range convs {
			if convIdx != 0 {
				convContent += "-----------------------------------\n"
			}
			convContent += fmt.Sprintf("%s\nM: %s\nQ: %s\nA: %s\n",
				util.FormatTime(conv.Time), conv.Mode, conv.Question, conv.Answer)
		}
		convCnt += len(convs)
	}
	convTitle := fmt.Sprintf("\n[%d convs | %d users]\n", convCnt, userCnt)
	body += convTitle + convContent

	email.SendEmail(subject, body)
}
