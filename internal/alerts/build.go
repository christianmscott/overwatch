package alerts

import (
	"github.com/christianmscott/overwatch/internal/alerts/smtp"
	"github.com/christianmscott/overwatch/internal/alerts/webhook"
	"github.com/christianmscott/overwatch/pkg/spec"
)

func BuildSenders(cfg spec.AlertsConfig) []AlertSender {
	var senders []AlertSender

	for _, w := range cfg.Webhooks {
		senders = append(senders, webhook.New(w))
	}

	if cfg.SMTP != nil {
		senders = append(senders, smtp.New(*cfg.SMTP))
	}

	return senders
}
