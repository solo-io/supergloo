package approval

import (
	"fmt"
	"sort"

	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

// For each VirtualMesh, sort them by accepted date, then invalidate if it applies to a Mesh that
// is already grouped into a VirtualMesh.
func validateOneVirtualMeshPerMesh(virtualMeshes []*networkingv1alpha2.VirtualMesh) {
	sortVirtualMeshesByAcceptedDate(virtualMeshes)

	vMeshesPerMesh := map[string]*networkingv1alpha2.VirtualMesh{}
	invalidVirtualMeshes := v1alpha2sets.NewVirtualMeshSet()

	// track accepted index
	var acceptedIndex uint32
	// Invalidate VirtualMesh if it applies to a Mesh that already has an applied VirtualMesh.
	for _, vMesh := range virtualMeshes {
		vMesh := vMesh
		for _, mesh := range vMesh.Spec.Meshes {
			// Ignore virtual mesh if previously invalidated.
			if invalidVirtualMeshes.Has(vMesh) {
				continue
			}
			meshKey := sets.Key(mesh)
			existingVirtualMesh, ok := vMeshesPerMesh[meshKey]
			if !ok {
				vMeshesPerMesh[meshKey] = vMesh
				vMesh.Status.State = networkingv1alpha2.ApprovalState_ACCEPTED
				acceptedIndex++
			} else {
				vMesh.Status = networkingv1alpha2.VirtualMeshStatus{
					ObservedGeneration: vMesh.Generation,
					State:              networkingv1alpha2.ApprovalState_INVALID,
					ValidationErrors: []string{fmt.Sprintf("Includes a Mesh (%s.%s) that already is grouped in a VirtualMesh (%s.%s)",
						mesh.Name, mesh.Namespace,
						existingVirtualMesh.Name, existingVirtualMesh.Namespace,
					)},
				}
				invalidVirtualMeshes.Insert(vMesh)
			}
		}
	}
}

// sort the set of virtual meshes in the order in which they were accepted.
// VMeshes which were accepted first and have not changed (i.e. their observedGeneration is up-to-date) take precedence.
// Next are vMeshes that were previously accepted but whose observedGeneration is out of date. This permits vmeshes which were modified but formerly correct to maintain
// their acceptance status ahead of vmeshes which were unmodified and previously rejected.
// Next will be the vmeshes which have been modified and rejected.
// Finally, vmeshes which are rejected and modified
func sortVirtualMeshesByAcceptedDate(virtualMeshes networkingv1alpha2.VirtualMeshSlice) {
	isUpToDate := func(vm *networkingv1alpha2.VirtualMesh) bool {
		return vm.Status.ObservedGeneration == vm.Generation
	}

	sort.SliceStable(virtualMeshes, func(i, j int) bool {
		vMesh1, vMesh2 := virtualMeshes[i], virtualMeshes[j]

		state1 := vMesh1.Status.State
		state2 := vMesh2.Status.State

		//// Accepted takes priority over Pending, which takes priority over Invalid
		//if state1 == networkingv1alpha2.ApprovalState_PENDING {
		//	return state2 != networkingv1alpha2.ApprovalState_ACCEPTED
		//} else if state2 == networkingv1alpha2.ApprovalState_PENDING {
		//	return state1 != networkingv1alpha2.ApprovalState_INVALID && state1 != networkingv1alpha2.ApprovalState_FAILED
		//}

		switch {
		case state1 == networkingv1alpha2.ApprovalState_ACCEPTED:
			if state2 != networkingv1alpha2.ApprovalState_ACCEPTED {
				// accepted comes before non accepted
				return true
			}

			if vMesh1UpToDate := isUpToDate(vMesh1); vMesh1UpToDate != isUpToDate(vMesh2) {
				// up to date is validated before modified
				return vMesh1UpToDate
			}

			return true
		case state2 == networkingv1alpha2.ApprovalState_ACCEPTED:
			// accepted comes before non accepted
			return false
		default:
			// neither policy has been accepted, we can simply sort by unique key
			return sets.Key(vMesh1) < sets.Key(vMesh2)
		}
	})
}
