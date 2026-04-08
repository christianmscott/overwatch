package checks

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
)

const certExpiryWarning = 7 * 24 * time.Hour // 7 days

type TLSChecker struct{}

func (t *TLSChecker) Check(ctx context.Context, check spec.CheckSpec) spec.CheckResult {
	start := time.Now()
	result := spec.CheckResult{
		CheckName: check.Name,
		Timestamp: start,
	}

	host, _, err := net.SplitHostPort(check.Target)
	if err != nil {
		host = check.Target
		check.Target = check.Target + ":443"
	}

	d := tls.Dialer{Config: &tls.Config{ServerName: host}}
	conn, err := d.DialContext(ctx, "tcp", check.Target)
	result.Duration = time.Since(start)
	if err != nil {
		result.Status = spec.StatusDown
		result.Error = err.Error()
		return result
	}
	defer conn.Close()

	tlsConn := conn.(*tls.Conn)
	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		result.Status = spec.StatusDown
		result.Error = "no certificates presented"
		return result
	}

	leaf := certs[0]
	until := time.Until(leaf.NotAfter)

	result.Detail = map[string]any{
		"subject":       leaf.Subject.CommonName,
		"issuer":        leaf.Issuer.CommonName,
		"expiresAt":     leaf.NotAfter.Format("2006-01-02"),
		"daysRemaining": int(until.Hours() / 24),
	}

	if until <= 0 {
		result.Status = spec.StatusDown
		result.Error = fmt.Sprintf("certificate expired %s ago", (-until).Round(time.Hour))
	} else if until < certExpiryWarning {
		result.Status = spec.StatusDegraded
		result.Error = fmt.Sprintf("certificate expires in %s", until.Round(time.Hour))
	} else {
		result.Status = spec.StatusUp
	}

	return result
}
