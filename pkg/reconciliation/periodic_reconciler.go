package reconciliation

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/rand"
)

func NewPeriodicReconciliationRunner() PeriodicReconciliationRunner {
	return &periodicReconciler{}
}

type periodicReconciler struct{}

func (p *periodicReconciler) Start(ctx context.Context, reconciliationPeriod time.Duration, reconciler Reconciler) {
	logger := contextutils.LoggerFrom(contextutils.WithLoggerValues(ctx,
		zap.String("periodic_reconciler_name", reconciler.GetName()),
	))

	if reconciliationPeriod <= 0 {
		panic("Cannot specify a non-positive reconciliation period")
	}

	// we want to either add or subtract a random number in the range [-10%, +10%] of the indicated reconciliation period
	// 10% isn't significant, just a nice number
	// this is to prevent a thundering herd scenario with our reconcilers https://en.wikipedia.org/wiki/Thundering_herd_problem
	tenPercentOfReconcilePeriod := reconciliationPeriod / 10
	timeFuzz := rand.Int63nRange(0, int64(tenPercentOfReconcilePeriod))

	shouldSubtractFuzz := rand.Int()%2 == 0
	if shouldSubtractFuzz {
		// 50% of the time we want to subtract 10%
		timeFuzz = -1 * timeFuzz
	}
	fuzzedTickerPeriod := reconciliationPeriod + time.Duration(timeFuzz)

	ticker := time.NewTicker(fuzzedTickerPeriod)
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
				logger.Errorf("Error during %s reconciliation, retrying in %s (fuzzed from %s): %+v", reconciler.GetName(), fuzzedTickerPeriod.String(), reconciliationPeriod.String(), err)
			}
		}
	}
}
