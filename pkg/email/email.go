package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/jekyulll/url_shortener/config"
	mail "github.com/jordan-wright/email"
)

type EmailSend struct {
	addr    string
	host    string
	myEail  string
	subject string
	auth    smtp.Auth
}

func NewEmailSend(cfg config.EmailConfig) (*EmailSend, error) {
	return &EmailSend{
		addr:    fmt.Sprintf("%s:%s", cfg.HostAddress, cfg.HostPort),
		host:    cfg.HostAddress,
		auth:    smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.HostAddress),
		myEail:  cfg.Username,
		subject: cfg.Subject,
	}, nil
}

func (e *EmailSend) Send(email string, emailCode string) error {
	instance := mail.NewEmail()
	instance.From = e.myEail
	instance.To = []string{email}
	instance.Subject = e.subject
	instance.Text = []byte(fmt.Sprintf("Your Verification code is: %s", emailCode))

	// 465 端口是隐式 SSL，必须用 SendWithTLS 直接建立 TLS 连接；
	// 普通的 Send 走明文 + STARTTLS，连 465 会直接失败
	return instance.SendWithTLS(e.addr, e.auth, &tls.Config{ServerName: e.host})
}
