package checks

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
)

type DNSChecker struct{}

func (d *DNSChecker) Check(ctx context.Context, check spec.CheckSpec) spec.CheckResult {
	start := time.Now()
	result := spec.CheckResult{
		CheckName: check.Name,
		Timestamp: start,
	}

	var resolver net.Resolver
	addrs, err := resolver.LookupHost(ctx, check.Target)
	result.Duration = time.Since(start)
	if err != nil {
		result.Status = spec.StatusDown
		result.Error = err.Error()
		return result
	}

	if len(addrs) == 0 {
		result.Status = spec.StatusDown
		result.Error = "no addresses returned"
		return result
	}

	result.Status = spec.StatusUp
	result.Detail = map[string]any{"resolved": strings.Join(addrs, ", ")}
	return result
}
