package reconciliation

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

func NewPeriodicReconciliationRunner() PeriodicReconciliationRunner {
	return &periodicReconciler{}
}

type periodicReconciler struct{}

func (p *periodicReconciler) Start(ctx context.Context, reconciliationPeriod time.Duration, reconciler Reconciler) {
	logger := contextutils.LoggerFrom(contextutils.WithLoggerValues(ctx,
		zap.String("periodic_reconciler_name", reconciler.GetName()),
	))

	ticker := time.NewTicker(reconciliationPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Debugf("periodic reconciler is ending", reconciler.GetName())
			return
		case <-ticker.C:
			logger.Debugf("new reconcile loop running")
			err := reconciler.Reconcile(ctx)
			if err != nil {
				logger.Errorf("Error during %s reconciliation, retrying in %s: %+v", reconciler.GetName(), reconciliationPeriod.String(), err)
			}
		}
	}
}
