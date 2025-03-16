package notify

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailConfig struct {
	To      []string
	Subject string
	Body    string
}

func SendEmail(config EmailConfig) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	headers := make(map[string]string)
	headers["From"] = smtpUser
	headers["To"] = config.To[0]
	headers["Subject"] = config.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + config.Body

	return smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		smtpUser,
		config.To,
		[]byte(message),
	)
}
