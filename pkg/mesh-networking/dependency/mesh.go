package dependency

import (
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

func markPendingTranslationsForMeshes(
	eventObjs []ezkube.ResourceId,
	meshes []*discoveryv1.Mesh,
) {
	for _, mesh := range meshes {
		markPendingTranslationsForMesh(eventObjs, mesh)
	}
}

func markPendingTranslationsForMesh(
	eventObjs []ezkube.ResourceId,
	mesh *discoveryv1.Mesh,
) {
	for _, eventObj := range eventObjs {
		// VirtualMesh
		if mesh.Status.GetAppliedVirtualMesh() != nil {
			markPendingVirtualMesh(eventObj, mesh, mesh.Status.GetAppliedVirtualMesh())
		}
	}
}

// Reprocess the applied TrafficPolicy if:
//  1. Mesh changes
//  2. VirtualMesh changes
func markPendingVirtualMesh(
	eventObj ezkube.ResourceId,
	mesh *discoveryv1.Mesh,
	appliedVirtualMesh *discoveryv1.MeshStatus_AppliedVirtualMesh,
) {
	var pending = false

	switch eventObj.(type) {
	case *discoveryv1.Mesh:
		if ezkube.RefsMatch(eventObj, mesh) {
			pending = true
		}
	case *networkingv1.VirtualMesh:
		if ezkube.RefsMatch(eventObj, appliedVirtualMesh.Ref) {
			pending = true
		}
	}

	appliedVirtualMesh.Pending = pending
}
