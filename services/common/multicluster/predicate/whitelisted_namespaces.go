package mc_predicate

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// a predicate that only allows an event to be processed if its namespace is included in arg namespaces
func WhitelistedNamespacePredicateProvider(set sets.String) predicate.Predicate {
	return &whitelistedNamespacesProvider{
		whitelistedNamespaces: set,
	}
}

type whitelistedNamespacesProvider struct {
	whitelistedNamespaces sets.String
}

func (p *whitelistedNamespacesProvider) Create(event event.CreateEvent) bool {
	return p.whitelistedNamespaces.Has(event.Meta.GetNamespace())
}

// Delete returns true if the Delete event should be processed
func (p *whitelistedNamespacesProvider) Delete(event event.DeleteEvent) bool {
	return p.whitelistedNamespaces.Has(event.Meta.GetNamespace())
}

// Update returns true if the Update event should be processed
func (p *whitelistedNamespacesProvider) Update(event event.UpdateEvent) bool {
	return p.whitelistedNamespaces.Has(event.MetaNew.GetNamespace()) && p.whitelistedNamespaces.Has(event.MetaOld.GetNamespace())
}

// Generic returns true if the Generic event should be processed
func (p *whitelistedNamespacesProvider) Generic(event event.GenericEvent) bool {
	return p.whitelistedNamespaces.Has(event.Meta.GetNamespace())
}
