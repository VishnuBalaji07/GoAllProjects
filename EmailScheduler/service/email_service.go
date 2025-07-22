package service

import (
	"log"
	"net/smtp"
)

type Mailer struct {
	From     string
	Password string
	Host     string
	Port     string
}

func (m *Mailer) Send(to, subject, body string) {
	auth := smtp.PlainAuth("", m.From, m.Password, m.Host)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	err := smtp.SendMail(m.Host+":"+m.Port, auth, m.From, []string{to}, msg)
	if err != nil {
		log.Println("Failed to send email:", err)
		return
	}
	log.Println("Email sent to", to)
}
