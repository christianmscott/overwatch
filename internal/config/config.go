package config

import (
	"fmt"
	"os"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
	"gopkg.in/yaml.v3"
)

const DefaultPath = "overwatch.yaml"

func Load(path string) (*spec.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg spec.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	applyDefaults(&cfg)

	if errs := Validate(&cfg); len(errs) > 0 {
		return nil, &ValidationError{Errors: errs}
	}

	return &cfg, nil
}

func applyDefaults(cfg *spec.Config) {
	for i := range cfg.Checks {
		if cfg.Checks[i].Timeout.Duration == 0 {
			cfg.Checks[i].Timeout = spec.Duration{Duration: 10 * time.Second}
		}
		if cfg.Checks[i].Interval.Duration == 0 {
			cfg.Checks[i].Interval = spec.Duration{Duration: 60 * time.Second}
		}
		if cfg.Checks[i].Type == spec.CheckHTTP && cfg.Checks[i].ExpectedStatus == 0 {
			cfg.Checks[i].ExpectedStatus = 200
		}
	}
	for i := range cfg.Alerts.Webhooks {
		if cfg.Alerts.Webhooks[i].Method == "" {
			cfg.Alerts.Webhooks[i].Method = "POST"
		}
		if cfg.Alerts.Webhooks[i].Timeout.Duration == 0 {
			cfg.Alerts.Webhooks[i].Timeout = spec.Duration{Duration: 10 * time.Second}
		}
	}
	if cfg.Worker.Concurrency == 0 {
		cfg.Worker.Concurrency = 4
	}
	if cfg.Worker.PollInterval.Duration == 0 {
		cfg.Worker.PollInterval = spec.Duration{Duration: 30 * time.Second}
	}
}
