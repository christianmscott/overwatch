package config

import (
	"fmt"
	"os"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
	"gopkg.in/yaml.v3"
)

const DefaultPath = "overwatch.yaml"

func Save(path string, cfg *spec.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

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
	if cfg.Server.BindAddress == "" {
		cfg.Server.BindAddress = "127.0.0.1"
	}
	if cfg.Server.BindPort == 0 {
		cfg.Server.BindPort = 3030
	}
	if cfg.Server.Concurrency == 0 {
		cfg.Server.Concurrency = 4
	}
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
		if cfg.Checks[i].Type == spec.CheckCheckIn && cfg.Checks[i].MaxSilence.Duration == 0 {
			cfg.Checks[i].MaxSilence = spec.Duration{Duration: 5 * time.Minute}
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
