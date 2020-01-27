package internal_watcher

import (
	"github.com/solo-io/mesh-projects/pkg/env"
	"github.com/solo-io/mesh-projects/services/common/multicluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// checks that a kubernetes object has the required metadata to be considered
// as a multicluster secret
func HasRequiredMetadata(metadata metav1.Object) bool {
	val, ok := metadata.GetLabels()[multicluster.MultiClusterLabel]
	return ok && val == "true" && metadata.GetNamespace() == env.DefaultWriteNamespace
}

// This object is used as a prerdicate interface for the secret event handler above
// The type is defined here: https://github.com/kubernetes-sigs/controller-runtime/blob/c14d8e600783ab7ca48c0d94f3b65daa2be92758/pkg/predicate/predicate.go#L27
type MultiClusterPredicate struct{}

func (m *MultiClusterPredicate) Create(e event.CreateEvent) bool {
	return HasRequiredMetadata(e.Meta)
}

func (m *MultiClusterPredicate) Delete(e event.DeleteEvent) bool {
	return HasRequiredMetadata(e.Meta)
}

func (m *MultiClusterPredicate) Update(e event.UpdateEvent) bool {
	return HasRequiredMetadata(e.MetaNew) || HasRequiredMetadata(e.MetaOld)
}

func (m *MultiClusterPredicate) Generic(e event.GenericEvent) bool {
	return HasRequiredMetadata(e.Meta)
}
