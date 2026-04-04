package checks

import (
	"context"
	"testing"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
)

func TestDNSCheckerUp(t *testing.T) {
	c := &DNSChecker{}
	result := c.Check(context.Background(), spec.CheckSpec{
		Name:    "test",
		Type:    spec.CheckDNS,
		Target:  "example.com",
		Timeout: spec.Duration{Duration: 5 * time.Second},
	})

	if result.Status != spec.StatusUp {
		t.Errorf("expected up, got %s: %s", result.Status, result.Error)
	}
}

func TestDNSCheckerDown(t *testing.T) {
	c := &DNSChecker{}
	result := c.Check(context.Background(), spec.CheckSpec{
		Name:    "test",
		Type:    spec.CheckDNS,
		Target:  "this-domain-does-not-exist-xyz123.invalid",
		Timeout: spec.Duration{Duration: 5 * time.Second},
	})

	if result.Status != spec.StatusDown {
		t.Errorf("expected down, got %s", result.Status)
	}
}
