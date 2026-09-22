package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// Mailer sends transactional email. When MAIL_HOST isn't configured, New
// returns a logMailer that just logs the message instead of failing outright —
// so invite-by-email still works locally (the invite link is always returned
// to the caller too, for manual sharing) without requiring SMTP credentials.
type Mailer interface {
	Send(to, subject, body string) error
	// IsLive reports whether this mailer actually delivers mail, as opposed to
	// just logging it.
	IsLive() bool
}

func New() Mailer {
	host := os.Getenv("MAIL_HOST")
	if host == "" {
		return &logMailer{}
	}
	return &smtpMailer{
		host:     host,
		port:     envOr("MAIL_PORT", "587"),
		username: os.Getenv("MAIL_USERNAME"),
		password: os.Getenv("MAIL_PASSWORD"),
		fromAddr: envOr("MAIL_FROM_ADDRESS", os.Getenv("MAIL_USERNAME")),
		fromName: envOr("MAIL_FROM_NAME", "DevSync"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type logMailer struct{}

func (m *logMailer) IsLive() bool { return false }

func (m *logMailer) Send(to, subject, body string) error {
	log.Printf("[mailer] MAIL_HOST not configured, logging instead of sending.\nTo: %s\nSubject: %s\n%s\n", to, subject, body)
	return nil
}

type smtpMailer struct {
	host, port, username, password, fromAddr, fromName string
}

func (m *smtpMailer) IsLive() bool { return true }

func (m *smtpMailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)

	msg := fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		m.fromName, m.fromAddr, to, subject, body,
	)

	// smtp.SendMail transparently upgrades to STARTTLS when the server
	// advertises it, which covers the common case (Gmail, most relays on :587).
	return smtp.SendMail(addr, auth, m.fromAddr, []string{to}, []byte(msg))
}
