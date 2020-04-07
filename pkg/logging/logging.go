package logging

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	LOG_LEVEL  = "LOG_LEVEL"
	DEBUG_MODE = "DEBUG_MODE"
)

const (
	EventTypeKey      = "event_type"
	ClusterNameKey    = "cluster_name"
	GroupVersion      = "group_version"
	Kind              = "kind"
	ResourceName      = "resource_name"
	ResourceNamespace = "resource_namespace"
	ResourceVersion   = "resource_version"
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
	gvk := obj.GetObjectKind().GroupVersionKind()
	ctx = contextutils.WithLoggerValues(ctx,
		zap.String(EventTypeKey, eventType.String()),
		zap.String(GroupVersion, gvk.GroupVersion().String()),
		zap.String(Kind, gvk.Kind),
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
	resourceVersion, err := accessor.ResourceVersion(obj)
	if err == nil {
		ctx = contextutils.WithLoggerValues(ctx, zap.String(ResourceVersion, resourceVersion))
	}

	return ctx
}

func BuildEventLogger(ctx context.Context, eventType EventType, obj runtime.Object) *zap.SugaredLogger {
	return contextutils.LoggerFrom(EventContext(ctx, eventType, obj))
}
