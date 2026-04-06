package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/christianmscott/overwatch/pkg/spec"
)

type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation failed:\n  - %s", strings.Join(e.Errors, "\n  - "))
}

func Validate(cfg *spec.Config) []string {
	var errs []string

	seen := make(map[string]bool)
	for i, c := range cfg.Checks {
		prefix := fmt.Sprintf("checks[%d]", i)

		if c.Name == "" {
			errs = append(errs, prefix+": name is required")
		} else if seen[c.Name] {
			errs = append(errs, prefix+": duplicate name "+c.Name)
		} else {
			seen[c.Name] = true
		}

		if c.Target == "" && c.Type != spec.CheckCheckIn {
			errs = append(errs, prefix+": target is required")
		}

		switch c.Type {
		case spec.CheckHTTP, spec.CheckTCP, spec.CheckTLS, spec.CheckDNS, spec.CheckCheckIn:
		default:
			errs = append(errs, fmt.Sprintf("%s: unknown type %q (want http, tcp, tls, dns, checkin)", prefix, c.Type))
		}

		if c.Type == spec.CheckCheckIn && c.MaxSilence.Duration == 0 {
			errs = append(errs, prefix+": max_silence is required for checkin checks")
		}

		if c.Type == spec.CheckHTTP && c.Target != "" {
			if _, err := url.ParseRequestURI(c.Target); err != nil {
				errs = append(errs, fmt.Sprintf("%s: invalid URL %q", prefix, c.Target))
			}
		}
	}

	for i, w := range cfg.Alerts.Webhooks {
		prefix := fmt.Sprintf("alerts.webhooks[%d]", i)
		if w.URL == "" {
			errs = append(errs, prefix+": url is required")
		} else if _, err := url.ParseRequestURI(w.URL); err != nil {
			errs = append(errs, fmt.Sprintf("%s: invalid url %q", prefix, w.URL))
		}
	}

	if smtp := cfg.Alerts.SMTP; smtp != nil {
		if smtp.Host == "" {
			errs = append(errs, "alerts.smtp: host is required")
		}
		if smtp.Port == 0 {
			errs = append(errs, "alerts.smtp: port is required")
		}
		if smtp.From == "" {
			errs = append(errs, "alerts.smtp: from is required")
		}
		if len(smtp.Recipients) == 0 {
			errs = append(errs, "alerts.smtp: at least one recipient is required")
		}
	}

	return errs
}
