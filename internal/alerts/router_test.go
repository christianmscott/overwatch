package alerts

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
)

type mockSender struct {
	mu       sync.Mutex
	messages []spec.AlertMessage
}

func (m *mockSender) Name() string { return "mock" }

func (m *mockSender) Send(_ context.Context, msg spec.AlertMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockSender) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.messages)
}

func (m *mockSender) last() spec.AlertMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.messages[len(m.messages)-1]
}

func TestRouterAlertsOnTransition(t *testing.T) {
	mock := &mockSender{}
	r := NewRouter([]AlertSender{mock})

	r.Handle(spec.CheckResult{CheckName: "a", Status: spec.StatusUp, Timestamp: time.Now()})
	if mock.count() != 0 {
		t.Fatalf("expected no alert on first result, got %d", mock.count())
	}

	r.Handle(spec.CheckResult{CheckName: "a", Status: spec.StatusUp, Timestamp: time.Now()})
	if mock.count() != 0 {
		t.Fatalf("expected no alert on same status, got %d", mock.count())
	}

	r.Handle(spec.CheckResult{CheckName: "a", Status: spec.StatusDown, Timestamp: time.Now()})
	if mock.count() != 1 {
		t.Fatalf("expected alert on up->down transition, got %d", mock.count())
	}

	last := mock.last()
	if last.PreviousStatus != spec.StatusUp {
		t.Errorf("expected previous status up, got %s", last.PreviousStatus)
	}
	if last.Status != spec.StatusDown {
		t.Errorf("expected status down, got %s", last.Status)
	}
}

func TestRouterSendTest(t *testing.T) {
	mock := &mockSender{}
	r := NewRouter([]AlertSender{mock})

	r.SendTest()
	if mock.count() != 1 {
		t.Fatalf("expected 1 test alert, got %d", mock.count())
	}
	if mock.last().CheckName != "test-alert" {
		t.Errorf("expected test-alert check name, got %s", mock.last().CheckName)
	}
}

func TestRouterMultipleSenders(t *testing.T) {
	m1 := &mockSender{}
	m2 := &mockSender{}
	r := NewRouter([]AlertSender{m1, m2})

	r.Handle(spec.CheckResult{CheckName: "a", Status: spec.StatusUp, Timestamp: time.Now()})
	r.Handle(spec.CheckResult{CheckName: "a", Status: spec.StatusDown, Timestamp: time.Now()})

	if m1.count() != 1 || m2.count() != 1 {
		t.Errorf("expected both senders to receive alert, got %d and %d", m1.count(), m2.count())
	}
}
