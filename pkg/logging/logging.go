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
		return "update"
	case DeleteEvent:
		return "delete"
	case GenericEvent:
		return "generic"
	default:
		return "unknown"
	}
}

func EventContext(ctx context.Context, eventType EventType, obj runtime.Object) context.Context {
	ctx = contextutils.WithLoggerValues(ctx,
		zap.String(EventTypeKey, eventType.String()),
		zap.String(GroupVersionKind, obj.GetObjectKind().GroupVersionKind().String()),
	)
	accessor := meta.NewAccessor()
	name, err := accessor.Name(obj)
	if err == nil {
		ctx = contextutils.WithLoggerValues(ctx, zap.String(ResourceName, name))
	}
	namespace, err := accessor.Namespace(obj)
	if err == nil {
		ctx = contextutils.WithLoggerValues(ctx, zap.String(ResourceNamespace, namespace))
	}
	return ctx
}

func BuildEventLogger(ctx context.Context, eventType EventType, obj runtime.Object) *zap.SugaredLogger {
	return contextutils.LoggerFrom(EventContext(ctx, eventType, obj))
}
