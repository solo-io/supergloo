package reconciliation

import (
	"context"
	"time"
)

type PeriodicReconciliationRunner interface {
	// If `reconciliationPeriod` == time.Duration(0), the method will panic.
	// Otherwise, this method blocks the current goroutine until <-ctx.Done() receives a value.
	// Any error returned from the reconciler will be logged, but will not cancel the reconciliation loop
	Start(ctx context.Context, reconciliationPeriod time.Duration, reconciler Reconciler)
}

type Reconciler interface {
	Reconcile(context.Context) error

	// for logger output
	GetName() string
}
