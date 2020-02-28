package csr_agent_controller

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type csrAgentPredicate struct {
	ctx context.Context
}

func CsrAgentPredicateProvider(ctx context.Context) predicate.Predicate {
	return &csrAgentPredicate{
		ctx: ctx,
	}
}

func (c *csrAgentPredicate) Create(event.CreateEvent) bool {
	return true
}

func (c *csrAgentPredicate) Delete(event.DeleteEvent) bool {
	return false
}

func (c *csrAgentPredicate) Update(updateEvent event.UpdateEvent) bool {
	newCsr, ok := updateEvent.ObjectNew.(*v1alpha1.MeshGroupCertificateSigningRequest)
	if !ok {
		contextutils.LoggerFrom(c.ctx).Warnf("Object %s.%s of type was passed into a predicate which only "+
			"accepts MeshGroups. GVK: %s", updateEvent.MetaNew.GetName(), updateEvent.MetaNew.GetNamespace(),
			updateEvent.ObjectNew.GetObjectKind().GroupVersionKind())
		return false
	}
	oldCsr, ok := updateEvent.ObjectOld.(*v1alpha1.MeshGroupCertificateSigningRequest)
	if !ok {
		contextutils.LoggerFrom(c.ctx).Warnf("Object %s.%s of type was passed into a predicate which only "+
			"accepts MeshGroups. GVK: %s", updateEvent.MetaOld.GetName(), updateEvent.MetaOld.GetNamespace(),
			updateEvent.ObjectOld.GetObjectKind().GroupVersionKind())
		return false
	}
	switch {
	case newCsr.Status.Equal(oldCsr.Status):
		return false
	case len(newCsr.Status.GetResponse().GetCaCertificate()) == 0 ||
		len(newCsr.Status.GetResponse().GetRootCertificate()) == 0:
		return false
	}
	// Only return true if the new certificate has been saturated
	return true
}

func (c *csrAgentPredicate) Generic(event.GenericEvent) bool {
	return false
}
