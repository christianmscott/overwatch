package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	gosmtp "net/smtp"
	"strings"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
)

type Sender struct {
	cfg spec.SMTPConfig
}

func New(cfg spec.SMTPConfig) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Name() string { return "smtp" }

func (s *Sender) Send(_ context.Context, msg spec.AlertMessage) error {
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))

	subject := fmt.Sprintf("[overwatch] %s is %s", msg.CheckName, msg.Status)
	body := formatBody(msg)
	mime := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		s.cfg.From,
		strings.Join(s.cfg.Recipients, ", "),
		subject,
		body,
	)

	var c *gosmtp.Client
	var err error

	if s.cfg.TLS {
		conn, dialErr := tls.DialWithDialer(
			&net.Dialer{Timeout: 10 * time.Second},
			"tcp", addr,
			&tls.Config{ServerName: s.cfg.Host},
		)
		if dialErr != nil {
			return fmt.Errorf("smtp: implicit TLS handshake failed on port %d (if your server uses STARTTLS, set tls: false and use port 587): %w", s.cfg.Port, dialErr)
		}
		c, err = gosmtp.NewClient(conn, s.cfg.Host)
	} else {
		c, err = gosmtp.Dial(addr)
		if err == nil {
			if ok, _ := c.Extension("STARTTLS"); ok {
				err = c.StartTLS(&tls.Config{ServerName: s.cfg.Host})
			}
		}
	}
	if err != nil {
		return fmt.Errorf("smtp connect: %w", err)
	}
	defer c.Close()

	if s.cfg.Username != "" {
		auth := gosmtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := c.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, rcpt := range s.cfg.Recipients {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt %s: %w", rcpt, err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(mime)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}

	return c.Quit()
}

func formatBody(msg spec.AlertMessage) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Check:    %s\n", msg.CheckName)
	fmt.Fprintf(&b, "Status:   %s\n", msg.Status)
	fmt.Fprintf(&b, "Previous: %s\n", msg.PreviousStatus)
	fmt.Fprintf(&b, "Time:     %s\n", msg.Timestamp.Format(time.RFC3339))
	if msg.Detail != "" {
		fmt.Fprintf(&b, "Detail:   %s\n", msg.Detail)
	}
	return b.String()
}
