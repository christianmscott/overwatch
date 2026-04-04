package alerts

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
)

type Router struct {
	senders []AlertSender

	mu    sync.Mutex
	state map[string]spec.CheckStatus
}

func NewRouter(senders []AlertSender) *Router {
	return &Router{
		senders: senders,
		state:   make(map[string]spec.CheckStatus),
	}
}

func (r *Router) Handle(result spec.CheckResult) {
	r.mu.Lock()
	prev, exists := r.state[result.CheckName]
	r.state[result.CheckName] = result.Status
	r.mu.Unlock()

	if !exists || prev == result.Status {
		return
	}

	msg := spec.AlertMessage{
		CheckName:      result.CheckName,
		Status:         result.Status,
		PreviousStatus: prev,
		Timestamp:      result.Timestamp,
		Detail:         result.Error,
	}

	r.dispatch(msg)
}

func (r *Router) SendTest() {
	msg := spec.AlertMessage{
		CheckName:      "test-alert",
		Status:         spec.StatusDown,
		PreviousStatus: spec.StatusUp,
		Timestamp:      time.Now(),
		Detail:         "This is a test alert from overwatch.",
	}
	r.dispatch(msg)
}

func (r *Router) dispatch(msg spec.AlertMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, s := range r.senders {
		if err := s.Send(ctx, msg); err != nil {
			slog.Error("alert send failed", "sender", s.Name(), "check", msg.CheckName, "error", err)
		} else {
			slog.Info("alert sent", "sender", s.Name(), "check", msg.CheckName, "status", msg.Status)
		}
	}
}
