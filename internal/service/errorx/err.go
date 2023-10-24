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

func RecordError(desc string, err error) {
	go func() {
		log.Println(desc, err)
		myErr := model.MyError{
			ErrorStr:           fmt.Sprintf("[%s]\n%s", desc, err.Error()),
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
	errCnt, errDesc := GetErrorsDesc(util.Today())
	subject := fmt.Sprintf("[%s/%s] Already %d Errors Today", util.GetAccount(), util.GetEnv(), errCnt)
	email.SendEmail(subject, errDesc)
}

func GetErrorsDesc(day string) (errCnt int, detail string) {
	errors, _ := store.GetErrors(day)
	errCnt = len(errors)
	for idx, myError := range errors {
		detail += util.TimestampToTimeStr(myError.TimestampInSeconds) + "\n" + myError.ErrorStr + "\n"
		if idx != errCnt-1 {
			detail += "-----------------------------------\n"
		}
	}
	return
}
