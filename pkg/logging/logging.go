package logging

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	EventTypeKey      = "event_type"
	ClusterNameKey    = "cluster_name"
	GroupVersionKind  = "group_version_kind"
	ResourceName      = "resource_name"
	ResourceNamespace = "resource_namespace"
)

type EventType int

const (
	CreateEvent EventType = iota
	UpdateEvent
	DeleteEvent
	GenericEvent
)

func (e EventType) String() string {
	switch e {
	case CreateEvent:
		return "create"
	case UpdateEvent:
		return "udpate"
	case DeleteEvent:
		return "delete"
	case GenericEvent:
		return "generic"
	default:
		return "unknown"
	}
}

func BuildEventLogger(ctx context.Context, eventType EventType, obj runtime.Object) *zap.SugaredLogger {
	logger := contextutils.LoggerFrom(ctx).With(
		zap.String(EventTypeKey, eventType.String()),
		zap.String(GroupVersionKind, obj.GetObjectKind().GroupVersionKind().String()),
	)

	accessor := meta.NewAccessor()
	name, err := accessor.Name(obj)
	if err == nil {
		logger = logger.With(ResourceName, name)
	}
	namespace, err := accessor.Namespace(obj)
	if err == nil {
		logger = logger.With(ResourceNamespace, namespace)
	}

	return logger
}
