package setup

import (
	"context"
	"os"

	"go.uber.org/zap"

	"github.com/solo-io/mesh-projects/pkg/env"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/stats"
)

func Main(ctx context.Context, errHandler func(error)) error {
	if os.Getenv("START_STATS_SERVER") != "" {
		stats.StartStatsServer()
	}

	writeNamespace := env.GetWriteNamespace()

	if errHandler == nil {
		errHandler = func(err error) {
			if err == nil {
				return
			}
			contextutils.LoggerFrom(ctx).Errorw("error", zap.Error(err))
		}
	}

	if err := Run(ctx, writeNamespace, errHandler); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
