package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/processfoundry/overwatch/pkg/spec"
)

type Sender struct {
	cfg spec.WebhookConfig
}

func New(cfg spec.WebhookConfig) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Name() string {
	if s.cfg.Name != "" {
		return s.cfg.Name
	}
	return "webhook"
}

func formatText(msg spec.AlertMessage) string {
	text := fmt.Sprintf("[%s] %s is %s (was %s)", msg.Timestamp.Format("15:04:05"), msg.CheckName, msg.Status, msg.PreviousStatus)
	if msg.Detail != "" {
		text += "\n" + msg.Detail
	}
	return text
}

func (s *Sender) Send(ctx context.Context, msg spec.AlertMessage) error {
	payload := map[string]interface{}{
		"text": formatText(msg),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal alert: %w", err)
	}

	method := s.cfg.Method
	if method == "" {
		method = http.MethodPost
	}

	req, err := http.NewRequestWithContext(ctx, method, s.cfg.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range s.cfg.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	if s.cfg.Timeout.Duration > 0 {
		client.Timeout = s.cfg.Timeout.Duration
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned %s", resp.Status)
	}

	return nil
}
