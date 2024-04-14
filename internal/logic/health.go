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
		body := buildBalanceSection(ohmygptBalance, sbBalance, api2dBalance)
		email.SendEmail("Insufficient Balance", body)
	}
	log.Println("Check finished")
}

func sendYesterdayReportEmail() {
	yesterday := util.Yesterday()
	subject := fmt.Sprintf("[%s/%s] Summary on %s", util.GetAccount(), util.GetEnv(), yesterday)

	body := ""
	ohmygptBalance, _ := ohmygpt.GetOhmygptBalance()
	sbBalance, _ := sb.GetSbBalance()
	api2dBalance, _ := api2d.GetApi2dBalance()
	balanceSect := buildBalanceSection(ohmygptBalance, sbBalance, api2dBalance)
	body += balanceSect

	if util.AccountIsBrother() {
		tradeNos, _ := store.GetSuccessOutTradeNos(yesterday)
		cnt := len(tradeNos)
		txnTitle := fmt.Sprintf("\n# %d purchases\n", cnt)
		txnContent := ""
		if cnt > 0 {
			txnContent = `
Time | User | Amount | Account
-----|------|--------|--------
`
			colTmpl := "%s | %s | %d | %s\n"
			for _, tradeNo := range tradeNos {
				transaction, _ := store.GetTransaction(tradeNo)
				paidAccount := constant.Brother
				openId := transaction.OpenId
				if transaction.UncleOpenId != "" {
					paidAccount = constant.Uncle
					openId = transaction.UncleOpenId
				}
				txnContent += fmt.Sprintf(colTmpl,
					util.FormatTime(transaction.UpdatedAt),
					openId,
					transaction.PriceInFen/100,
					paidAccount,
				)
			}
		}
		body += txnTitle + txnContent
	}

	errCnt, errorContent := errorx.GetErrorsDesc(yesterday)
	errorTitle := fmt.Sprintf("\n# %d errors\n", errCnt)
	body += errorTitle + errorContent

	users, _ := store.GetActiveUsers(yesterday)
	userCnt := len(users)
	convCnt := 0
	convContent := ""
	for idx, user := range users {
		convContent += fmt.Sprintf("## %d %s\n", idx+1, user)
		convs, _ := store.GetConversations(user, yesterday)
		for _, conv := range convs {
			convTmpl := `
### %s %s %d
**Q**: %s
**A**: %s
`
			convContent += fmt.Sprintf(convTmpl,
				util.FormatTime(conv.Time),
				conv.Mode,
				conv.PaidBalance,
				util.TruncateAndEscapeNewLine(conv.Question, 100),
				util.TruncateAndEscapeNewLine(conv.Answer, 100),
			)
		}
		convCnt += len(convs)
	}
	convTitle := fmt.Sprintf("\n# %d users | %d convs\n", userCnt, convCnt)
	body += convTitle + convContent

	email.SendEmail(subject, body)
}

func buildBalanceSection(ohmygptBalance float64, sbBalance float64, api2dBalance float64) string {
	balanceTmpl := `
# Balance
Vendor | Balance
-------|------
%s | %.2f
%s | %.2f
%s | %.2f
`
	balanceSect := fmt.Sprintf(balanceTmpl,
		"Ohmygpt", ohmygptBalance,
		"SB", sbBalance,
		"Api2d", api2dBalance,
	)
	return balanceSect
}
