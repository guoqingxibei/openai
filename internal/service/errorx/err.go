package errorx

import (
	"fmt"
	"log"
	"openai/internal/model"
	"openai/internal/service/email"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

func RecordError(title string, err error) {
	go func() {
		log.Println(title, err)
		myErr := model.MyError{
			Title:  title,
			Detail: err.Error(),
			Time:   time.Now(),
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
	errCnt, errDesc := GetErrorsDesc(util.Today())
	subject := fmt.Sprintf("[%s/%s] Already %d Errors Today", util.GetAccount(), util.GetEnv(), errCnt)
	email.SendEmail(subject, errDesc)
}

func GetErrorsDesc(day string) (errCnt int, detail string) {
	errors, _ := store.GetErrors(day)
	errCnt = len(errors)
	// reverse errors
	for idx, myError := range errors {
		if idx != 0 {
			detail = "-----------------------------------\n" + detail
		}
		detail = fmt.Sprintf("%s\n%s\n%s\n", util.FormatTime(myError.Time), myError.Title, myError.Detail) + detail
	}
	return
}
