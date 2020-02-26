package strategies

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/snapshot"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	MeshMetadataMissing = func(meshName string) error {
		return eris.Errorf("Mesh %s does not have complete metadata provided to federation strategy", meshName)
	}
)

type MeshMetadata struct {
	MeshWorkloads []*discovery_v1alpha1.MeshWorkload
	MeshServices  []*discovery_v1alpha1.MeshService
	ClusterName   string
}

type MeshNameToMetadata map[string]*MeshMetadata

// mesh name to the associated resources
type PerMeshMetadata struct {
	MeshNameToMetadata MeshNameToMetadata

	// all groups included here will have all their relevant data populated above
	// i.e., if a group is included here, you can safely query the map above for its member meshes' data
	ResolvedMeshGroups []*networking_v1alpha1.MeshGroup
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
	MeshGroup *networking_v1alpha1.MeshGroup
	Err       error
}

func BuildPerMeshMetadataFromSnapshot(ctx context.Context, snapshot snapshot.MeshNetworkingSnapshot, meshClient discovery_core.MeshClient) (PerMeshMetadata, []ErrorReport) {
	var errors []ErrorReport

	perMeshResources := PerMeshMetadata{
		MeshNameToMetadata: map[string]*MeshMetadata{},
	}

	// set up `meshNameToWorkloads`
	for _, workload := range snapshot.CurrentState.MeshWorkloads {
		meshName := workload.Spec.GetMesh().GetName()
		meshResources := perMeshResources.GetOrInitialize(meshName)

		meshResources.MeshWorkloads = append(meshResources.MeshWorkloads, workload)
	}

	// set up `meshNameToServices`
	for _, service := range snapshot.CurrentState.MeshServices {
		meshName := service.Spec.GetMesh().GetName()
		meshResources := perMeshResources.GetOrInitialize(meshName)

		meshResources.MeshServices = append(meshResources.MeshServices, service)
	}

	// set up `meshNameToClusterName`
	for _, group := range snapshot.CurrentState.MeshGroups {
		var multiErr *multierror.Error

		for _, memberMesh := range group.Spec.Meshes {
			resourcesForMesh := perMeshResources.GetOrInitialize(memberMesh.GetName())

			if resourcesForMesh.ClusterName != "" {
				// we've already found the cluster name for this mesh
				continue
			}

			meshObj, err := meshClient.Get(ctx, client.ObjectKey{
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
			perMeshResources.ResolvedMeshGroups = append(perMeshResources.ResolvedMeshGroups, group)
		} else {
			errors = append(errors, ErrorReport{
				MeshGroup: group,
				Err:       multiErr.ErrorOrNil(),
			})
		}
	}

	return perMeshResources, errors
}
