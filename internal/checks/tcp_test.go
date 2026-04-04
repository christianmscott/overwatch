package checks

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/christianmscott/overwatch/pkg/spec"
)

func TestTCPCheckerUp(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	c := &TCPChecker{}
	result := c.Check(context.Background(), spec.CheckSpec{
		Name:    "test",
		Type:    spec.CheckTCP,
		Target:  ln.Addr().String(),
		Timeout: spec.Duration{Duration: 5 * time.Second},
	})

	if result.Status != spec.StatusUp {
		t.Errorf("expected up, got %s: %s", result.Status, result.Error)
	}
}

func TestTCPCheckerDown(t *testing.T) {
	c := &TCPChecker{}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result := c.Check(ctx, spec.CheckSpec{
		Name:   "test",
		Type:   spec.CheckTCP,
		Target: "127.0.0.1:1", // port 1 is almost certainly not listening
	})

	if result.Status != spec.StatusDown {
		t.Errorf("expected down, got %s", result.Status)
	}
}
