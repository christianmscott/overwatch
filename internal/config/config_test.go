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
	if err := os.WriteFile(path, []byte(StarterConfig), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(cfg.Checks))
	}
	if cfg.Checks[0].Name != "example-http" {
		t.Errorf("expected check name example-http, got %s", cfg.Checks[0].Name)
	}
	if cfg.Checks[0].Type != spec.CheckHTTP {
		t.Errorf("expected check type http, got %s", cfg.Checks[0].Type)
	}
	if cfg.Worker.Concurrency != 4 {
		t.Errorf("expected concurrency 4, got %d", cfg.Worker.Concurrency)
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
	path := filepath.Join(dir, "empty.yaml")
	if err := os.WriteFile(path, []byte("checks: []\n"), 0644); err != nil {
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
