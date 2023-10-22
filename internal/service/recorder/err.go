package recorder

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
			ErrorStr:           fmt.Sprintf("[%s] %s", desc, err.Error()),
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
	subject := fmt.Sprintf("[%s/%s] Already %d Errors Today", util.GetAccount(), util.GetEnv(), count)
	email.SendEmail(subject, body)
}
