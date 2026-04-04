package checks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
)

func TestHTTPCheckerUp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := &HTTPChecker{}
	result := c.Check(context.Background(), spec.CheckSpec{
		Name:           "test",
		Type:           spec.CheckHTTP,
		Target:         srv.URL,
		Timeout:        spec.Duration{Duration: 5 * time.Second},
		ExpectedStatus: 200,
	})

	if result.Status != spec.StatusUp {
		t.Errorf("expected up, got %s: %s", result.Status, result.Error)
	}
}

func TestHTTPCheckerDown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	c := &HTTPChecker{}
	result := c.Check(context.Background(), spec.CheckSpec{
		Name:           "test",
		Type:           spec.CheckHTTP,
		Target:         srv.URL,
		Timeout:        spec.Duration{Duration: 5 * time.Second},
		ExpectedStatus: 200,
	})

	if result.Status != spec.StatusDown {
		t.Errorf("expected down, got %s", result.Status)
	}
}

func TestHTTPCheckerHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "test-value" {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c := &HTTPChecker{}
	result := c.Check(context.Background(), spec.CheckSpec{
		Name:           "test",
		Type:           spec.CheckHTTP,
		Target:         srv.URL,
		Timeout:        spec.Duration{Duration: 5 * time.Second},
		ExpectedStatus: 200,
		Headers:        map[string]string{"X-Custom": "test-value"},
	})

	if result.Status != spec.StatusUp {
		t.Errorf("expected up, got %s: %s", result.Status, result.Error)
	}
}

func TestHTTPCheckerTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	c := &HTTPChecker{}
	result := c.Check(ctx, spec.CheckSpec{
		Name:   "test",
		Type:   spec.CheckHTTP,
		Target: srv.URL,
	})

	if result.Status != spec.StatusDown {
		t.Errorf("expected down on timeout, got %s", result.Status)
	}
}
