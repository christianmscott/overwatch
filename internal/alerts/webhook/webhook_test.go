package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/processfoundry/overwatch/pkg/spec"
)

func TestSendSuccess(t *testing.T) {
	var received map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	s := New(spec.WebhookConfig{
		Name:    "test",
		URL:     srv.URL,
		Timeout: spec.Duration{Duration: 5 * time.Second},
	})

	msg := spec.AlertMessage{
		CheckName:      "my-check",
		Status:         spec.StatusDown,
		PreviousStatus: spec.StatusUp,
		Timestamp:      time.Now(),
		Detail:         "connection refused",
	}

	if err := s.Send(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text, ok := received["text"].(string)
	if !ok || text == "" {
		t.Error("expected non-empty text field in payload")
	}
	if !strings.Contains(text, "my-check") {
		t.Errorf("expected text to contain check name, got %s", text)
	}
}

func TestSendServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	s := New(spec.WebhookConfig{URL: srv.URL})
	err := s.Send(context.Background(), spec.AlertMessage{})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestSendCustomHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("missing custom header")
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	s := New(spec.WebhookConfig{
		URL:     srv.URL,
		Headers: map[string]string{"Authorization": "Bearer secret"},
	})

	if err := s.Send(context.Background(), spec.AlertMessage{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestName(t *testing.T) {
	s := New(spec.WebhookConfig{Name: "slack"})
	if s.Name() != "slack" {
		t.Errorf("expected slack, got %s", s.Name())
	}

	s2 := New(spec.WebhookConfig{})
	if s2.Name() != "webhook" {
		t.Errorf("expected webhook, got %s", s2.Name())
	}
}
