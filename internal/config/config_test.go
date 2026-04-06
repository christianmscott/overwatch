package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/christianmscott/overwatch/pkg/spec"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overwatch.yaml")
	content := `server:
  bind_address: 127.0.0.1
  bind_port: 3030
checks:
  - name: test-http
    type: http
    target: https://example.com
    interval: 60s
    timeout: 10s
alerts:
  webhooks: []
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(cfg.Checks))
	}
	if cfg.Checks[0].Name != "test-http" {
		t.Errorf("expected check name test-http, got %s", cfg.Checks[0].Name)
	}
	if cfg.Checks[0].Type != spec.CheckHTTP {
		t.Errorf("expected check type http, got %s", cfg.Checks[0].Type)
	}
}

func TestLoadStarterConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "overwatch.yaml")
	if err := os.WriteFile(path, []byte(StarterConfig), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("starter config should load without error: %v", err)
	}
	if len(cfg.Checks) != 0 {
		t.Fatalf("starter config should have 0 active checks, got %d", len(cfg.Checks))
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("checks:\n  - name: [broken"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadValidationErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	content := `checks:
  - name: ""
    type: bogus
    target: ""
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if len(ve.Errors) == 0 {
		t.Fatal("expected at least one validation error")
	}
}
