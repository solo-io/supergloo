package mc_predicate

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	KubeBlacklistedNamespaces = sets.NewString(
		"kube-system",
		"istio-system",
		"istio-operator",
	)
)

// a predicate that only allows an event to be processed if its namespace is NOT included in arg namespaces
func BlacklistedNamespacePredicateProvider(set sets.String) predicate.Predicate {
	return &blacklistedNamespacedProvider{
		blacklistedNamespaces: set,
	}
}

type blacklistedNamespacedProvider struct {
	blacklistedNamespaces sets.String
}

func (p *blacklistedNamespacedProvider) Create(event event.CreateEvent) bool {
	return !p.blacklistedNamespaces.Has(event.Meta.GetNamespace())
}

// Delete returns true if the Delete event should be processed
func (p *blacklistedNamespacedProvider) Delete(event event.DeleteEvent) bool {
	return !p.blacklistedNamespaces.Has(event.Meta.GetNamespace())
}

// Update returns true if the Update event should be processed
func (p *blacklistedNamespacedProvider) Update(event event.UpdateEvent) bool {
	return !p.blacklistedNamespaces.Has(event.MetaNew.GetNamespace()) && !p.blacklistedNamespaces.Has(event.MetaOld.GetNamespace())
}

// Generic returns true if the Generic event should be processed
func (p *blacklistedNamespacedProvider) Generic(event event.GenericEvent) bool {
	return !p.blacklistedNamespaces.Has(event.Meta.GetNamespace())
}
