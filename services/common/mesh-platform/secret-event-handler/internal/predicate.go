package internal_watcher

import (
	"strings"

	"github.com/solo-io/service-mesh-hub/pkg/env"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/mesh-platform/k8s"
	k8s_core_types "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

const (
	SoloRegistrationSecretPrefix = "solo.io/register/"
)

// checks that a kubernetes object has the required metadata to be considered
// as a multicluster secret
func HasRequiredMetadata(metadata metav1.Object, secret runtime.Object) bool {
	// TODO Choose a consistent secret filtering approach once migrated to skv2
	val, ok := metadata.GetLabels()[mc_manager.MultiClusterLabel]
	return ((ok && val == "true") || strings.HasPrefix(string(secret.(*k8s_core_types.Secret).Type), SoloRegistrationSecretPrefix)) && metadata.GetNamespace() == env.GetWriteNamespace()
}

// This object is used as a prerdicate interface for the secret event handler above
// The type is defined here: https://github.com/kubernetes-sigs/controller-runtime/blob/c14d8e600783ab7ca48c0d94f3b65daa2be92758/pkg/predicate/predicate.go#L27
type MultiClusterPredicate struct{}

func (m *MultiClusterPredicate) Create(e event.CreateEvent) bool {
	return HasRequiredMetadata(e.Meta, e.Object)
}

func (m *MultiClusterPredicate) Delete(e event.DeleteEvent) bool {
	return HasRequiredMetadata(e.Meta, e.Object)
}

func (m *MultiClusterPredicate) Update(e event.UpdateEvent) bool {
	return HasRequiredMetadata(e.MetaNew, e.ObjectNew) || HasRequiredMetadata(e.MetaOld, e.ObjectOld)
}

func (m *MultiClusterPredicate) Generic(e event.GenericEvent) bool {
	return HasRequiredMetadata(e.Meta, e.Object)
}
