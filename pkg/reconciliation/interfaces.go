package reconciliation

import (
	"context"
	"time"
)

type PeriodicReconciler interface {
	// If `reconciliationPeriod` == time.Duration(0), the method will panic.
	// Otherwise, this method blocks the current goroutine until <-ctx.Done() receives a value.
	// Any error returned from `action` will be logged, but will not cancel the reconciliation loop
	Start(ctx context.Context, reconciliationPeriod time.Duration, actionName string, action func(context.Context) error)
}
