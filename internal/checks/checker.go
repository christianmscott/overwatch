package checks

import (
	"context"
	"fmt"
	"time"

	"github.com/processfoundry/overwatch/pkg/spec"
)

type Checker interface {
	Check(ctx context.Context, check spec.CheckSpec) spec.CheckResult
}

var registry = map[spec.CheckType]Checker{}

func Register(t spec.CheckType, c Checker) {
	registry[t] = c
}

func Get(t spec.CheckType) (Checker, error) {
	c, ok := registry[t]
	if !ok {
		return nil, fmt.Errorf("unknown check type: %s", t)
	}
	return c, nil
}

func Run(ctx context.Context, check spec.CheckSpec) spec.CheckResult {
	c, err := Get(check.Type)
	if err != nil {
		return spec.CheckResult{
			CheckName: check.Name,
			Status:    spec.StatusUnknown,
			Error:     err.Error(),
			Timestamp: time.Now(),
		}
	}

	ctx, cancel := context.WithTimeout(ctx, check.Timeout.Duration)
	defer cancel()

	return c.Check(ctx, check)
}

func init() {
	Register(spec.CheckHTTP, &HTTPChecker{})
	Register(spec.CheckTCP, &TCPChecker{})
	Register(spec.CheckTLS, &TLSChecker{})
	Register(spec.CheckDNS, &DNSChecker{})
	Register(spec.CheckCheckIn, DefaultCheckIn)
}
