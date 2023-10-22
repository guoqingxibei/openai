package email

import (
	"fmt"
	"log"
	"net/smtp"
	"openai/internal/config"
)

var emailConfig = config.C.Email

func SendEmail(subject string, body string) {
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
