package spec

import "time"

type CheckStatus string

const (
	StatusUp       CheckStatus = "up"
	StatusDown     CheckStatus = "down"
	StatusDegraded CheckStatus = "degraded"
	StatusUnknown  CheckStatus = "unknown"
)

type CheckType string

const (
	CheckHTTP    CheckType = "http"
	CheckTCP     CheckType = "tcp"
	CheckTLS     CheckType = "tls"
	CheckDNS     CheckType = "dns"
	CheckCheckIn CheckType = "checkin"
)

type CheckSpec struct {
	Name           string            `yaml:"name" json:"name"`
	Type           CheckType         `yaml:"type" json:"type"`
	Target         string            `yaml:"target,omitempty" json:"target,omitempty"`
	Interval       Duration          `yaml:"interval" json:"interval"`
	Timeout        Duration          `yaml:"timeout" json:"timeout"`
	Headers        map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	ExpectedStatus int               `yaml:"expected_status,omitempty" json:"expected_status,omitempty"`
	MaxSilence     Duration          `yaml:"max_silence,omitempty" json:"max_silence,omitempty"`
	Alerts         []string          `yaml:"alerts,omitempty" json:"alerts,omitempty"`
}

type CheckResult struct {
	CheckName string        `json:"check_name"`
	Status    CheckStatus   `json:"status"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}
