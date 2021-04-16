package dependency

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// Only run translation relevant for the changed eventObjs by marking the applied policies that require translation.
func MarkPendingTranslations(
	eventObjs []ezkube.ResourceId,
	destinations []*discoveryv1.Destination,
	meshes []*discoveryv1.Mesh,
) {
	markPendingTranslationsForDestinations(eventObjs, destinations)
	markPendingTranslationsForMeshes(eventObjs, meshes)
}

// Loop through every object and clear any pending flags
func ClearPendingTranslation(
	destinations []*discoveryv1.Destination,
	meshes []*discoveryv1.Mesh,
) {
	for _, destination := range destinations {
		for _, appliedTrafficPolicy := range destination.Status.GetAppliedTrafficPolicies() {
			appliedTrafficPolicy.Pending = false
		}
		for _, appliedAccessPolicy := range destination.Status.GetAppliedAccessPolicies() {
			appliedAccessPolicy.Pending = false
		}
		if destination.Status.GetAppliedFederation() != nil {
			destination.Status.GetAppliedFederation().Pending = false
		}
	}

	for _, mesh := range meshes {
		if mesh.Status.GetAppliedVirtualMesh() != nil {
			mesh.Status.GetAppliedVirtualMesh().Pending = false
		}
	}
}
