package email

import (
	"fmt"
	"log"
	"net/smtp"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/store"
	"time"
)

func init() {
	go func() {
		for true {
			time.Sleep(time.Second * 10)
			SendEmail("email notification test", "this is a test")
		}
	}()
}

var emailConfig = config.C.Email

func SendEmail(subject string, body string) {
	status, _ := store.GetEmailNotificationStatus()
	if status == constant.Off {
		return
	}

	smtpServer := emailConfig.SmtpServer
	from := emailConfig.From
	to := emailConfig.To
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s",
		from, to, subject, body)
	err := smtp.SendMail(smtpServer+":587",
		smtp.PlainAuth("", from, emailConfig.Pass, smtpServer),
		from,
		[]string{to},
		[]byte(msg),
	)
	if err != nil {
		log.Println("smtp.SendMail() failed", err)
		return
	}
}
