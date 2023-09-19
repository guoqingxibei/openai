package logic

import (
	"fmt"
	"github.com/robfig/cron"
	"log"
	"openai/internal/model"
	"openai/internal/service/api2d"
	"openai/internal/service/email"
	"openai/internal/service/gptredis"
	"openai/internal/service/sb"
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
	point, _ := api2d.GetPoint()
	if point < 1000 {
		alarm = true
	}
	balance, _ := sb.GetSbBalance()
	if balance < 5 {
		alarm = true
	}
	if alarm {
		log.Println("Balance is insufficient, sending email...")
		body := fmt.Sprintf("Api2d points is %d and SB balance is %.2f", point, balance)
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
		_ = gptredis.AppendError(today, myErr)
		errCount, _ := gptredis.GetErrorsLen(today)
		if errCount%1 == 0 {
			sendErrorAlarmEmail()
		}
	}()
}

func sendErrorAlarmEmail() {
	errors, _ := gptredis.GetErrors(util.Today())
	body := ""
	for _, myError := range errors {
		body += util.TimestampToTimeStr(myError.TimestampInSeconds) + "  " + myError.ErrorStr + "\n"
	}
	subject := fmt.Sprintf("[%s]Already %d ChatAPI Errors Today", util.GetAccount(), len(errors))
	email.SendEmail(subject, body)
}

func sendYesterdayReportEmail() {
	yesterday := util.Yesterday()
	subject := fmt.Sprintf("[%s]Summary for %s", util.GetAccount(), yesterday)

	body := ""
	errorContent := ""
	errors, _ := gptredis.GetErrors(yesterday)
	for _, myError := range errors {
		errorContent += util.TimestampToTimeStr(myError.TimestampInSeconds) + "  " + myError.ErrorStr + "\n"
	}
	errorTitle := fmt.Sprintf("[%d errors in total]\n", len(errors))
	body += errorTitle + errorContent

	point, _ := api2d.GetPoint()
	api2dTitle := "\n[Api2d]\n"
	api2dContent := fmt.Sprintf("There are %d points left\n", point)
	body += api2dTitle + api2dContent

	balance, _ := sb.GetSbBalance()
	sbTitle := "\n[SB]\n"
	sbContent := fmt.Sprintf("There are %.2f RMB left", balance)
	body += sbTitle + sbContent
	email.SendEmail(subject, body)
}
