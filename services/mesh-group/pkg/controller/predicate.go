package controller

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

/*
	IgnoreStatusUpdatePredicate exists so that the mesh group operator does not handle update events when only the
	status of the object has changed. The operator will often update the status of the object it is processing,
	and we do not want that update event to propogate back into the operator.

	For every other event type it always returns true
*/
func IgnoreStatusUpdatePredicate(ctx context.Context) predicate.Predicate {
	return &ignoreStatusPredicate{
		ctx: ctx,
	}
}

type ignoreStatusPredicate struct {
	ctx context.Context
}

// Create returns true if the Delete event should be processed
func (p *ignoreStatusPredicate) Create(event event.CreateEvent) bool {
	return true
}

// Delete returns true if the Delete event should be processed
func (p *ignoreStatusPredicate) Delete(event event.DeleteEvent) bool {
	return true
}

// Update returns true if the Update event should be processed
func (p *ignoreStatusPredicate) Update(event event.UpdateEvent) bool {
	logger := contextutils.LoggerFrom(p.ctx)
	oldObject, ok := event.ObjectOld.(*v1alpha1.MeshGroup)
	if !ok {
		// Appending the link as it is the only field on the meta which will contain the type of the object passed in
		logger.Warnf("Object %s.%s of type was passed into a predicate which only accepts MeshGroups. link %s",
			event.MetaOld.GetName(), event.MetaOld.GetNamespace(), event.MetaOld.GetSelfLink())
		// this should never happen, if accidentally applied to the wrong controller, let it through
		return true
	}
	newObject, ok := event.ObjectNew.(*v1alpha1.MeshGroup)
	if !ok {
		// Appending the link as it is the only field on the meta which will contain the type of the object passed in
		logger.Warnf("Object %s.%s of type was passed into a predicate which only accepts MeshGroups. link %s",
			event.MetaOld.GetName(), event.MetaOld.GetNamespace(), event.MetaOld.GetSelfLink())
		// this should never happen, if accidentally applied to the wrong controller, let it through
		return true
	}
	return oldObject.Status.Equal(newObject.Status)
}

// Generic returns true if the Generic event should be processed
func (p *ignoreStatusPredicate) Generic(event event.GenericEvent) bool {
	return true
}
