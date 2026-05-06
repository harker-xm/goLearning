package email

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

// SMTPInfo holds the SMTP server configuration.
type SMTPInfo struct {
	Host     string
	Port     int
	IsSSL    bool
	UserName string
	Password string
	From     string
}

// Email wraps SMTP configuration for sending emails.
type Email struct {
	*SMTPInfo
}

// NewEmail creates a new Email instance with the given SMTP config.
func NewEmail(info *SMTPInfo) *Email {
	return &Email{SMTPInfo: info}
}

// SendMail sends an HTML email to the given recipients.
func (e *Email) SendMail(to []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	dialer := gomail.NewDialer(e.Host, e.Port, e.UserName, e.Password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: e.IsSSL}
	return dialer.DialAndSend(m)
}
