package alerts

import (
	"context"

	"github.com/processfoundry/overwatch/pkg/spec"
)

type AlertSender interface {
	Send(ctx context.Context, msg spec.AlertMessage) error
	Name() string
}
