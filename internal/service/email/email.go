package email

import (
	"fmt"
	"log"
	"net/smtp"
	"openai/internal/config"
	"openai/internal/constant"
	"openai/internal/store"
	"openai/internal/util"
)

var emailConfig = config.C.Email

func SendEmail(subject string, mdBody string) {
	htmlBody := util.MarkdownToHtml(mdBody)
	status, _ := store.GetEmailNotificationStatus()
	if status == constant.Off {
		return
	}

	smtpServer := emailConfig.SmtpServer
	from := emailConfig.From
	to := emailConfig.To
	tmpl := `From: %s
To: %s
Subject: %s
MIME-version: 1.0;
Content-Type: text/html; 

%s
`
	msg := fmt.Sprintf(tmpl,
		from, to, subject, htmlBody)
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
