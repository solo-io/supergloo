package multicluster

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	restrictedNamespaces = sets.NewString(
		"kube-system",
	)
)

// a predicate that only allows an event to be processed if its namespace is NOT included in `restrictedNamespaces`
func RestrictedNamespacePredicate() predicate.Predicate {
	return &namespaceDeploymentPredicate{
		restrictedNamespaces: restrictedNamespaces,
	}
}

type namespaceDeploymentPredicate struct {
	restrictedNamespaces sets.String
}

func (p *namespaceDeploymentPredicate) Create(event event.CreateEvent) bool {
	return !p.restrictedNamespaces.Has(event.Meta.GetNamespace())
}

// Delete returns true if the Delete event should be processed
func (p *namespaceDeploymentPredicate) Delete(event event.DeleteEvent) bool {
	return !p.restrictedNamespaces.Has(event.Meta.GetNamespace())
}

// Update returns true if the Update event should be processed
func (p *namespaceDeploymentPredicate) Update(event event.UpdateEvent) bool {
	return !p.restrictedNamespaces.Has(event.MetaNew.GetNamespace()) && !p.restrictedNamespaces.Has(event.MetaOld.GetNamespace())
}

// Generic returns true if the Generic event should be processed
func (p *namespaceDeploymentPredicate) Generic(event event.GenericEvent) bool {
	return !p.restrictedNamespaces.Has(event.Meta.GetNamespace())
}
