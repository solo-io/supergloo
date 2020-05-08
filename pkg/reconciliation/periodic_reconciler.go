package reconciliation

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

func NewPeriodicReconciler() PeriodicReconciler {
	return &periodicReconciler{}
}

type periodicReconciler struct{}

func (p *periodicReconciler) Start(ctx context.Context, reconciliationPeriod time.Duration, actionName string, action func(context.Context) error) {
	logger := contextutils.LoggerFrom(contextutils.WithLoggerValues(ctx,
		zap.String("periodic_reconciler_name", actionName),
	))

	ticker := time.NewTicker(reconciliationPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Debugf("periodic reconciler is ending", actionName)
			return
		case <-ticker.C:
			logger.Debugf("new reconcile loop running")
			err := action(ctx)
			if err != nil {
				logger.Errorf("Error during reconciliation, retrying in %s: %+v", reconciliationPeriod.String(), err)
			}
		}
	}
}
