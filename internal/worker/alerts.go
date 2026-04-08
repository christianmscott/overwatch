package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/christianmscott/overwatch/internal/alerts/smtp"
	"github.com/christianmscott/overwatch/internal/alerts/webhook"
	"github.com/christianmscott/overwatch/pkg/spec"
)

type txQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// dispatchAlerts reads bound alert channels for the monitor and fires them.
// Returns whether any alert was triggered and a summary of outcomes.
func dispatchAlerts(ctx context.Context, tx txQuerier, lease spec.Lease, result spec.CheckResult, prevStatus string) (bool, any) {
	msg := spec.AlertMessage{
		CheckName:      result.CheckName,
		Status:         result.Status,
		PreviousStatus: spec.CheckStatus(prevStatus),
		Timestamp:      result.Timestamp,
		Detail:         result.Error,
	}

	rows, err := tx.Query(ctx, `
		SELECT ac.id, ac.type, ac.config
		FROM "AlertChannel" ac
		JOIN "MonitorAlertBinding" b ON b."alertChannelId" = ac.id
		WHERE b."monitorId" = $1 AND ac.enabled = true
	`, lease.MonitorID)
	if err != nil {
		slog.Error("alert: query channels", "monitor", lease.MonitorID, "error", err)
		return false, nil
	}
	defer rows.Close()

	type outcome struct {
		Channel string `json:"channel"`
		Status  string `json:"status"`
		Error   string `json:"error,omitempty"`
	}
	var outcomes []outcome
	triggered := false

	for rows.Next() {
		var channelID, channelType string
		var rawConfig []byte
		if err := rows.Scan(&channelID, &channelType, &rawConfig); err != nil {
			continue
		}

		sender, err := buildSender(channelType, rawConfig)
		if err != nil {
			slog.Warn("alert: unsupported channel type", "type", channelType)
			continue
		}

		triggered = true
		dispatchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		sendErr := sender.Send(dispatchCtx, msg)
		cancel()

		o := outcome{Channel: channelID, Status: "sent"}
		if sendErr != nil {
			o.Status = "failed"
			o.Error = sendErr.Error()
			slog.Error("alert send failed", "channel", channelID, "type", channelType, "error", sendErr)
		} else {
			slog.Info("alert sent", "channel", channelID, "type", channelType, "check", msg.CheckName)
		}
		outcomes = append(outcomes, o)
	}

	if len(outcomes) == 0 {
		return triggered, nil
	}
	return triggered, outcomes
}

type alertSender interface {
	Send(ctx context.Context, msg spec.AlertMessage) error
}

func buildSender(channelType string, rawConfig []byte) (alertSender, error) {
	var cfg map[string]any
	if err := json.Unmarshal(rawConfig, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal channel config: %w", err)
	}

	switch channelType {
	case "WEBHOOK":
		url := cfgStr(cfg, "webhookUrl", "")
		if url == "" {
			return nil, fmt.Errorf("webhook: missing webhookUrl")
		}
		return webhook.New(spec.WebhookConfig{Name: "webhook", URL: url}), nil

	case "EMAIL":
		recipient := cfgStr(cfg, "email", "")
		if recipient == "" {
			return nil, fmt.Errorf("email: missing email address")
		}
		smtpCfg, err := smtpConfigFromEnv(recipient)
		if err != nil {
			return nil, err
		}
		return smtp.New(smtpCfg), nil

	default:
		return nil, fmt.Errorf("unsupported channel type %q", channelType)
	}
}

// smtpConfigFromEnv builds an SMTPConfig from environment variables.
// Required: SMTP_HOST, SMTP_FROM. Optional: SMTP_PORT (default 587), SMTP_USERNAME, SMTP_PASSWORD, SMTP_TLS.
func smtpConfigFromEnv(recipient string) (spec.SMTPConfig, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return spec.SMTPConfig{}, fmt.Errorf("email: SMTP_HOST env var not set")
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		return spec.SMTPConfig{}, fmt.Errorf("email: SMTP_FROM env var not set")
	}

	port := 587
	if p := os.Getenv("SMTP_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	useTLS := os.Getenv("SMTP_TLS") == "true"

	return spec.SMTPConfig{
		Host:       host,
		Port:       port,
		Username:   os.Getenv("SMTP_USERNAME"),
		Password:   os.Getenv("SMTP_PASSWORD"),
		From:       from,
		Recipients: []string{recipient},
		TLS:        useTLS,
	}, nil
}
