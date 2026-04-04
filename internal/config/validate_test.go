package config

import (
	"strings"
	"testing"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
)

func validConfig() *spec.Config {
	return &spec.Config{
		Checks: []spec.CheckSpec{
			{
				Name:     "test-http",
				Type:     spec.CheckHTTP,
				Target:   "https://example.com",
				Interval: spec.Duration{Duration: 60 * time.Second},
				Timeout:  spec.Duration{Duration: 10 * time.Second},
			},
		},
	}
}

func TestValidateValid(t *testing.T) {
	errs := Validate(validConfig())
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestValidateNoChecks(t *testing.T) {
	cfg := &spec.Config{}
	errs := Validate(cfg)
	if len(errs) == 0 {
		t.Fatal("expected error for no checks")
	}
	if !containsStr(errs, "at least one check") {
		t.Errorf("expected 'at least one check' error, got %v", errs)
	}
}

func TestValidateDuplicateNames(t *testing.T) {
	cfg := validConfig()
	cfg.Checks = append(cfg.Checks, cfg.Checks[0])
	errs := Validate(cfg)
	if !containsStr(errs, "duplicate name") {
		t.Errorf("expected duplicate name error, got %v", errs)
	}
}

func TestValidateMissingTarget(t *testing.T) {
	cfg := validConfig()
	cfg.Checks[0].Target = ""
	errs := Validate(cfg)
	if !containsStr(errs, "target is required") {
		t.Errorf("expected target required error, got %v", errs)
	}
}

func TestValidateUnknownType(t *testing.T) {
	cfg := validConfig()
	cfg.Checks[0].Type = "ftp"
	errs := Validate(cfg)
	if !containsStr(errs, "unknown type") {
		t.Errorf("expected unknown type error, got %v", errs)
	}
}

func TestValidateWebhookMissingURL(t *testing.T) {
	cfg := validConfig()
	cfg.Alerts.Webhooks = []spec.WebhookConfig{{Name: "test"}}
	errs := Validate(cfg)
	if !containsStr(errs, "url is required") {
		t.Errorf("expected url required error, got %v", errs)
	}
}

func TestValidateSMTPMissingFields(t *testing.T) {
	cfg := validConfig()
	cfg.Alerts.SMTP = &spec.SMTPConfig{}
	errs := Validate(cfg)
	if !containsStr(errs, "host is required") {
		t.Errorf("expected smtp host error, got %v", errs)
	}
	if !containsStr(errs, "port is required") {
		t.Errorf("expected smtp port error, got %v", errs)
	}
}

func containsStr(ss []string, substr string) bool {
	for _, s := range ss {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
