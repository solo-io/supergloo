package logging

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

type EventType int

const (
	CreateEvent EventType = iota
	UpdateEvent
	DeleteEvent
	GenericEvent
)

func BuildEventLogger(ctx context.Context, eventType EventType, clusterName string) *zap.SugaredLogger {
	logger := contextutils.LoggerFrom(ctx).With("cluster_name", clusterName)
	switch eventType {
	case CreateEvent:
		logger = logger.With("event_type", "create")
	case UpdateEvent:
		logger = logger.With("event_type", "update")
	case DeleteEvent:
		logger = logger.With("event_type", "delete")
	case GenericEvent:
		logger = logger.With("event_type", "generic")
	default:
		logger = logger.With("event_type", "unknown")
	}
	return logger
}
