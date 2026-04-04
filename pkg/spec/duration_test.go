package spec

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestDurationUnmarshal(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{`"30s"`, 30 * time.Second},
		{`"5m"`, 5 * time.Minute},
		{`"1h"`, time.Hour},
		{`"100ms"`, 100 * time.Millisecond},
	}
	for _, tt := range tests {
		var d Duration
		if err := yaml.Unmarshal([]byte(tt.input), &d); err != nil {
			t.Errorf("unmarshal %s: %v", tt.input, err)
			continue
		}
		if d.Duration != tt.expected {
			t.Errorf("unmarshal %s: got %v, want %v", tt.input, d.Duration, tt.expected)
		}
	}
}

func TestDurationUnmarshalInvalid(t *testing.T) {
	var d Duration
	if err := yaml.Unmarshal([]byte(`"not-a-duration"`), &d); err == nil {
		t.Error("expected error for invalid duration")
	}
}

func TestDurationMarshal(t *testing.T) {
	d := Duration{Duration: 30 * time.Second}
	out, err := yaml.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "30s\n" {
		t.Errorf("got %q, want %q", string(out), "30s\n")
	}
}
