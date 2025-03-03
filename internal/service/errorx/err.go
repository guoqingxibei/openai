package errorx

import (
	"fmt"
	"log/slog"
	"openai/internal/model"
	"openai/internal/service/email"
	"openai/internal/store"
	"openai/internal/util"
	"time"
)

func RecordError(title string, err error) {
	RecordErrorWithEmailOption(title, err, true)
}

func RecordErrorWithEmailOption(title string, err error, sendEmail bool) {
	go func() {
		slog.Error(title, "error", err)
		if !util.EnvIsProd() || !sendEmail {
			return
		}

		myErr := model.MyError{
			Account: util.GetAccount(),
			Title:   title,
			Detail:  err.Error(),
			Time:    time.Now(),
		}
		_ = store.AppendError(util.Today(), myErr)
		sendErrorAlarmEmail()
	}()
}

func sendErrorAlarmEmail() {
	errCnt, errDesc := GetErrorsDesc(util.Today())
	subject := fmt.Sprintf("Already %d Errors Today", errCnt)
	email.SendEmail(subject, errDesc)
}

func GetErrorsDesc(day string) (errCnt int, detail string) {
	errors, _ := store.GetErrors(day)
	errCnt = len(errors)
	// reverse errors
	for _, myError := range errors {
		errorTmpl := `
### %s %s
**%s**
%s
`
		detail = fmt.Sprintf(
			errorTmpl,
			util.FormatTime(myError.Time),
			myError.Account,
			myError.Title,
			myError.Detail) + detail
	}
	return
}
