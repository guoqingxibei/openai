package logic

import (
	"fmt"
	"github.com/robfig/cron"
	"log/slog"
	"openai/internal/constant"
	"openai/internal/model"
	"openai/internal/service/email"
	"openai/internal/service/errorx"
	"openai/internal/service/ohmygpt"
	"openai/internal/store"
	"openai/internal/util"
)

func init() {
	if !util.AccountIsBrother() || !util.EnvIsProd() {
		return
	}

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

func checkVendorBalance() {
	slog.Info("Checking balance of vendors...")
	alarm := false
	ohmygptBalance, _ := ohmygpt.GetOhmygptBalance()
	if ohmygptBalance < 5 {
		alarm = true
	}
	if alarm {
		slog.Info("Balance is insufficient, sending email...")
		body := buildBalanceSection(ohmygptBalance)
		email.SendEmail("Insufficient Balance", body)
	}
	slog.Info("Check finished")
}

func sendYesterdayReportEmail() {
	yesterday := util.Yesterday()
	subject := fmt.Sprintf("Summary on %s", yesterday)
	body := ""

	// balance section
	ohmygptBalance, _ := ohmygpt.GetOhmygptBalance()
	balanceSect := buildBalanceSection(ohmygptBalance)
	body += balanceSect

	// transaction section
	txnTitle := "\n# Order\n"
	tradeNos, _ := store.GetSuccessOutTradeNos(yesterday)
	cnt := len(tradeNos)
	txnContent := fmt.Sprintf("## %d orders in total\n", cnt)
	if cnt > 0 {
		txnContent += `
Time | Account | Amount
:---:|:-------:|:-----:
`
		txnColTmpl := "%s | %s | %.1f\n"
		for _, tradeNo := range tradeNos {
			transaction, _ := store.GetTransaction(tradeNo)
			paidAccount := constant.Brother
			if transaction.UncleOpenId != "" {
				paidAccount = constant.Uncle
			}
			txnContent += fmt.Sprintf(txnColTmpl,
				util.FormatTime(transaction.UpdatedAt),
				paidAccount,
				float64(transaction.PriceInFen)/100,
			)
		}
	}
	body += txnTitle + txnContent

	// error section
	errorTitle := "\n# Error\n"
	errCnt, errorDesc := errorx.GetErrorsDesc(yesterday)
	errorContent := fmt.Sprintf("## %d errors in total\n", errCnt) + errorDesc
	body += errorTitle + errorContent

	// usage section
	brotherUsers, _ := store.GetActiveUsers(yesterday, false)
	var brotherConvs []model.Conversation
	for _, user := range brotherUsers {
		convs, _ := store.GetConversations(user, yesterday, false)
		brotherConvs = append(brotherConvs, convs...)
	}
	uncleUsers, _ := store.GetActiveUsers(yesterday, true)
	var uncleConvs []model.Conversation
	for _, user := range uncleUsers {
		convs, _ := store.GetConversations(user, yesterday, true)
		uncleConvs = append(uncleConvs, convs...)
	}
	usageTitle := "\n# Usage\n"
	usageContent := `
Account | Users | Conversations
:------:|:-----:|:------------:
`
	usageColTmpl := "%s | %d | %d\n"
	usageContent += fmt.Sprintf(usageColTmpl, "brother", len(brotherUsers), len(brotherConvs))
	usageContent += fmt.Sprintf(usageColTmpl, "uncle", len(uncleUsers), len(uncleConvs))
	usageContent += fmt.Sprintf(
		usageColTmpl,
		"TOTAL",
		len(brotherUsers)+len(uncleUsers),
		len(brotherConvs)+len(uncleConvs),
	)
	body += usageTitle + usageContent

	email.SendEmail(subject, body)
}

func buildBalanceSection(ohmygptBalance float64) string {
	balanceTmpl := `
# Balance
Vendor | Balance
:-----:|:-------:
%s | %.2f
`
	balanceSect := fmt.Sprintf(balanceTmpl,
		"Ohmygpt", ohmygptBalance,
	)
	return balanceSect
}
