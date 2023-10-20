package logic

import (
	"fmt"
	"github.com/robfig/cron"
	"log"
	"openai/internal/model"
	"openai/internal/service/api2d"
	"openai/internal/service/email"
	"openai/internal/service/ohmygpt"
	"openai/internal/service/sb"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

func init() {
	c1 := cron.New()
	// Execute once every day at 00:00
	err := c1.AddFunc("0 0 0 * * ?", func() {
		sendYesterdayReportEmail()
	})
	if err != nil {
		log.Println("AddFunc failed:", err)
		return
	}
	c1.Start()

	if util.AccountIsUncle() {
		c2 := cron.New()
		// Execute once every half an hour
		err = c2.AddFunc("0 */10 * * * *", func() {
			checkVendorBalance()
		})
		if err != nil {
			log.Println("AddFunc failed:", err)
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
	if sbBalance < 10 {
		alarm = true
	}
	api2dBalance, _ := api2d.GetApi2dBalance()
	if api2dBalance < 10 {
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

func RecordError(err error) {
	go func() {
		myErr := model.MyError{
			ErrorStr:           err.Error(),
			TimestampInSeconds: time.Now().Unix(),
		}
		today := util.Today()
		_ = store.AppendError(today, myErr)
		errCount, _ := store.GetErrorsLen(today)
		if errCount%1 == 0 {
			sendErrorAlarmEmail()
		}
	}()
}

func sendErrorAlarmEmail() {
	errors, _ := store.GetErrors(util.Today())
	count := len(errors)
	body := ""
	for idx, myError := range errors {
		body += util.TimestampToTimeStr(myError.TimestampInSeconds) + "  " + myError.ErrorStr + "\n"
		if idx != count-1 {
			body += "-----------------------------------\n"
		}
	}
	subject := fmt.Sprintf("[%s] Already %d Errors Today", util.GetAccount(), count)
	email.SendEmail(subject, body)
}

func sendYesterdayReportEmail() {
	yesterday := util.Yesterday()
	subject := fmt.Sprintf("[%s] Summary for %s", util.GetAccount(), yesterday)

	body := ""
	errorContent := ""
	errors, _ := store.GetErrors(yesterday)
	count := len(errors)
	for idx, myError := range errors {
		errorContent += util.TimestampToTimeStr(myError.TimestampInSeconds) + "  " + myError.ErrorStr + "\n"
		if idx != count-1 {
			body += "-----------------------------------\n"
		}
	}
	errorTitle := fmt.Sprintf("[%d errors]\n", count)
	body += errorTitle + errorContent

	ohmygptBalance, _ := ohmygpt.GetOhmygptBalance()
	sbBalance, _ := sb.GetSbBalance()
	api2dBalance, _ := api2d.GetApi2dBalance()
	balanceTitle := "\n[balance]\n"
	balanceContent := fmt.Sprintf("Ohmygpt: ￥%.2f\nSB: ￥%.2f\nApi2d: ￥%.2f",
		ohmygptBalance, sbBalance, api2dBalance)
	body += balanceTitle + balanceContent

	email.SendEmail(subject, body)
}
