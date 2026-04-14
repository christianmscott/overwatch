package checks

import (
	"context"
	"testing"

	"github.com/processfoundry/overwatch/pkg/spec"
)

func TestRunUnknownType(t *testing.T) {
	result := Run(context.Background(), spec.CheckSpec{
		Name: "test",
		Type: "ftp",
	})
	if result.Status != spec.StatusUnknown {
		t.Errorf("expected unknown, got %s", result.Status)
	}
}

func TestRegistryContainsAllTypes(t *testing.T) {
	for _, ct := range []spec.CheckType{spec.CheckHTTP, spec.CheckTCP, spec.CheckTLS, spec.CheckDNS} {
		if _, err := Get(ct); err != nil {
			t.Errorf("expected checker for %s, got error: %v", ct, err)
		}
	}
}
