package strategies

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	networking_snapshot "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/snapshot"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	MeshMetadataMissing = func(meshName string) error {
		return eris.Errorf("Mesh %s does not have complete metadata provided to federation strategy", meshName)
	}
)

type MeshMetadata struct {
	MeshWorkloads []*smh_discovery.MeshWorkload
	MeshServices  []*smh_discovery.MeshService
	ClusterName   string
}

type MeshNameToMetadata map[string]*MeshMetadata

// mesh name to the associated resources
type PerMeshMetadata struct {
	MeshNameToMetadata MeshNameToMetadata

	// all virtual meshes included here will have all their relevant data populated above
	// i.e., if a virtual mesh is included here, you can safely query the map above for its member meshes' data
	ResolvedVirtualMeshes []*smh_networking.VirtualMesh
}

func (p PerMeshMetadata) GetOrInitialize(meshName string) *MeshMetadata {
	resources, ok := p.MeshNameToMetadata[meshName]
	if !ok {
		empty := &MeshMetadata{}
		p.MeshNameToMetadata[meshName] = empty
		return empty
	}

	return resources
}

type ErrorReport struct {
	VirtualMesh *smh_networking.VirtualMesh
	Err         error
}

func BuildPerMeshMetadataFromSnapshot(
	ctx context.Context,
	snapshot *networking_snapshot.MeshNetworkingSnapshot,
	meshClient smh_discovery.MeshClient,
) (PerMeshMetadata, []ErrorReport) {
	var errors []ErrorReport

	perMeshResources := PerMeshMetadata{
		MeshNameToMetadata: map[string]*MeshMetadata{},
	}

	// set up `meshNameToWorkloads`
	for _, workload := range snapshot.MeshWorkloads {
		meshName := workload.Spec.GetMesh().GetName()
		meshResources := perMeshResources.GetOrInitialize(meshName)

		meshResources.MeshWorkloads = append(meshResources.MeshWorkloads, workload)
	}

	// set up `meshNameToServices`
	for _, service := range snapshot.MeshServices {
		meshName := service.Spec.GetMesh().GetName()
		meshResources := perMeshResources.GetOrInitialize(meshName)

		meshResources.MeshServices = append(meshResources.MeshServices, service)
	}

	// set up `meshNameToClusterName`
	for _, vm := range snapshot.VirtualMeshes {
		var multiErr *multierror.Error

		for _, memberMesh := range vm.Spec.Meshes {
			resourcesForMesh := perMeshResources.GetOrInitialize(memberMesh.GetName())

			if resourcesForMesh.ClusterName != "" {
				// we've already found the cluster name for this mesh
				continue
			}

			meshObj, err := meshClient.GetMesh(ctx, client.ObjectKey{
				Name:      memberMesh.GetName(),
				Namespace: memberMesh.GetNamespace(),
			})
			if err != nil {
				multiErr = multierror.Append(multiErr, err)
				continue
			}

			perMeshResources.GetOrInitialize(memberMesh.GetName()).ClusterName = meshObj.Spec.GetCluster().GetName()
		}

		if multiErr.ErrorOrNil() == nil {
			perMeshResources.ResolvedVirtualMeshes = append(perMeshResources.ResolvedVirtualMeshes, vm)
		} else {
			errors = append(errors, ErrorReport{
				VirtualMesh: vm,
				Err:         multiErr.ErrorOrNil(),
			})
		}
	}

	return perMeshResources, errors
}
