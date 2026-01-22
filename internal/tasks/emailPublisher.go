package tasks

import (
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/NurulloMahmud/medicalka-project/config"
)

type EmailSender struct {
	cfg    config.SMTP
	logger *log.Logger
}

func NewEmailSender(cfg config.SMTP, logger *log.Logger) *EmailSender {
	return &EmailSender{
		cfg:    cfg,
		logger: logger,
	}
}

func (e *EmailSender) SendVerificationEmail(to, token string) {
	go e.sendWithRetry(to, token, 3, time.Minute)
}

func (e *EmailSender) sendWithRetry(to, token string, maxRetries int, interval time.Duration) {
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = e.send(to, token)
		if err == nil {
			e.logger.Printf("verification email sent to %s", to)
			return
		}

		e.logger.Printf("attempt %d/%d failed for %s: %v", attempt, maxRetries, to, err)

		if attempt < maxRetries {
			time.Sleep(interval)
		}
	}

	e.logger.Printf("all %d attempts failed for %s: %v", maxRetries, to, err)
}

func (e *EmailSender) send(to, token string) error {
	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)

	subject := "Verify your email"
	verifyURL := fmt.Sprintf("http://localhost:8080/auth/verify-email?token=%s", token)
	body := fmt.Sprintf("Click the link to verify your email: %s", verifyURL)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.cfg.From, to, subject, body,
	))

	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)
	return smtp.SendMail(addr, auth, e.cfg.From, []string{to}, msg)
}
