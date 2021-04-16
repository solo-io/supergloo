package dependency

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func markPendingTranslationsForDestinations(
	eventObjs []ezkube.ResourceId,
	destinations []*discoveryv1.Destination,
) {
	for _, destination := range destinations {
		markPendingTranslationsForDestination(eventObjs, destination)
	}
}

func markPendingTranslationsForDestination(
	eventObjs []ezkube.ResourceId,
	destination *discoveryv1.Destination,
) {
	for _, eventObj := range eventObjs {
		// TrafficPolicies
		for _, appliedTrafficPolicy := range destination.Status.GetAppliedTrafficPolicies() {
			markPendingTrafficPolicies(eventObj, destination, appliedTrafficPolicy)
		}
		// AccessPolicies
		for _, appliedAccessPolicy := range destination.Status.GetAppliedAccessPolicies() {
			markPendingAccessPolicies(eventObj, destination, appliedAccessPolicy)
		}
		// Federation
		if destination.Status.GetAppliedFederation() != nil {
			markPendingFederation(eventObj, destination, destination.Status.GetAppliedFederation())
		}

	}
}

// Reprocess the applied TrafficPolicy if:
//  1. Destination changes
//  2. TrafficPolicy changes
func markPendingTrafficPolicies(
	eventObj ezkube.ResourceId,
	destination *discoveryv1.Destination,
	appliedTrafficPolicy *discoveryv1.DestinationStatus_AppliedTrafficPolicy,
) {
	var pending = false

	switch eventObj.(type) {
	case *discoveryv1.Destination:
		if ezkube.RefsMatch(eventObj, destination) {
			pending = true
		}
	case *networkingv1.TrafficPolicy:
		if ezkube.RefsMatch(eventObj, appliedTrafficPolicy.Ref) {
			pending = true
		}
	}

	appliedTrafficPolicy.Pending = pending
}

// Reprocess the applied TrafficPolicy if:
//  1. Destination changes
//  2. AccessPolicy changes
func markPendingAccessPolicies(
	eventObj ezkube.ResourceId,
	destination *discoveryv1.Destination,
	appliedAccessPolicy *discoveryv1.DestinationStatus_AppliedAccessPolicy,
) {
	var pending = false

	switch eventObj.(type) {
	case *discoveryv1.Destination:
		if ezkube.RefsMatch(eventObj, destination) {
			pending = true
		}
	case *networkingv1.AccessPolicy:
		if ezkube.RefsMatch(eventObj, appliedAccessPolicy.Ref) {
			pending = true
		}
	}

	appliedAccessPolicy.Pending = pending
}

// Reprocess the applied TrafficPolicy if:
//  1. Destination changes
//  2. AccessPolicy changes
func markPendingFederation(
	eventObj ezkube.ResourceId,
	destination *discoveryv1.Destination,
	appliedFederation *discoveryv1.DestinationStatus_AppliedFederation,
) {
	var pending = false

	switch eventObj.(type) {
	case *discoveryv1.Destination:
		if ezkube.RefsMatch(eventObj, destination) {
			pending = true
		}
	case *networkingv1.VirtualMesh:
		if ezkube.RefsMatch(eventObj, appliedFederation.VirtualMeshRef) {
			pending = true
		}
	}

	appliedFederation.Pending = pending
}
