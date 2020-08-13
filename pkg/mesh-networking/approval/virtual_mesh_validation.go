package approval

import (
	"fmt"
	"sort"

	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/skv2/contrib/pkg/sets"
)

// For each VirtualMesh, sort them by accepted date, then invalidate if it applies to a Mesh that
// is already grouped into a VirtualMesh.
func validateOneVirtualMeshPerMesh(virtualMeshes []*networkingv1alpha2.VirtualMesh) {
	sortVirtualMeshesByAcceptedDate(virtualMeshes)

	vMeshesPerMesh := map[string]*networkingv1alpha2.VirtualMesh{}

	// track accepted index
	var acceptedIndex uint32
	// Invalidate VirtualMesh if it applies to a Mesh that already has an applied VirtualMesh.
	for _, vMesh := range virtualMeshes {
		vMesh := vMesh
		for _, mesh := range vMesh.Spec.Meshes {
			meshKey := sets.Key(mesh)
			existingVirtualMesh, ok := vMeshesPerMesh[meshKey]
			if !ok {
				vMeshesPerMesh[meshKey] = vMesh
				vMesh.Status.Status = &networkingv1alpha2.ApprovalStatus{
					AcceptanceOrder: acceptedIndex,
					State:           networkingv1alpha2.ApprovalState_ACCEPTED,
				}
				acceptedIndex++
			} else {
				vMesh.Status = networkingv1alpha2.VirtualMeshStatus{
					ObservedGeneration: vMesh.Generation,
					Status: &networkingv1alpha2.ApprovalStatus{
						State: networkingv1alpha2.ApprovalState_INVALID,
						Errors: []string{fmt.Sprintf("Applies to a Mesh that already is grouped in a VirtualMesh %s.%s",
							existingVirtualMesh.Name,
							existingVirtualMesh.Namespace)},
					},
				}
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

		status1 := vMesh1.Status.Status
		status2 := vMesh2.Status.Status

		// Accepted takes priority over Pending (nil), which takes priority over Invalid
		if status1 == nil {
			return status2 == nil || status2.State == networkingv1alpha2.ApprovalState_INVALID || status2.State == networkingv1alpha2.ApprovalState_FAILED
		} else if status2 == nil {
			return status1.State != networkingv1alpha2.ApprovalState_INVALID && status1.State != networkingv1alpha2.ApprovalState_FAILED
		}

		switch {
		case status1.State == networkingv1alpha2.ApprovalState_ACCEPTED:
			if status2.State != networkingv1alpha2.ApprovalState_ACCEPTED {
				// accepted comes before non accepted
				return true
			}

			if vMesh1UpToDate := isUpToDate(vMesh1); vMesh1UpToDate != isUpToDate(vMesh2) {
				// up to date is validated before modified
				return vMesh1UpToDate
			}

			// sort by the previous acceptance order
			return status1.AcceptanceOrder < status2.AcceptanceOrder
		case status2.State == networkingv1alpha2.ApprovalState_ACCEPTED:
			// accepted comes before non accepted
			return false
		default:
			// neither policy has been accepted, we can simply sort by unique key
			return sets.Key(vMesh1) < sets.Key(vMesh2)
		}
	})
}
