package strategies

import (
	"context"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/federation/dns"
)

func NewPermissiveFederation(meshServiceClient smh_discovery.MeshServiceClient) FederationStrategy {
	return &permissiveFederation{
		meshServiceClient: meshServiceClient,
	}
}

type permissiveFederation struct {
	meshServiceClient smh_discovery.MeshServiceClient
}

func (p *permissiveFederation) WriteFederationToServices(
	ctx context.Context,
	vm *smh_networking.VirtualMesh,
	meshNameToMetadata MeshNameToMetadata,
) error {
	for _, serverMeshRef := range vm.Spec.Meshes {
		serverMeshMetadata, ok := meshNameToMetadata[serverMeshRef.GetName()]
		if !ok {
			return MeshMetadataMissing(serverMeshRef.GetName())
		}

		servicesInMesh := serverMeshMetadata.MeshServices

		var federatedToWorkloads []*smh_core_types.ResourceRef
		for _, clientMesh := range vm.Spec.Meshes {
			// skip `serverMeshRef` - we don't want to federate a service to the same mesh that it's in
			if clientMesh.GetName() == serverMeshRef.GetName() && clientMesh.GetNamespace() == serverMeshRef.GetNamespace() {
				continue
			}

			clientMeshMetadata, ok := meshNameToMetadata[clientMesh.GetName()]
			if !ok {
				return MeshMetadataMissing(clientMesh.GetName())
			}

			// get the workloads belonging to this mesh (the mesh that the clients are in)
			for _, workload := range clientMeshMetadata.MeshWorkloads {
				federatedToWorkloads = append(federatedToWorkloads, &smh_core_types.ResourceRef{
					Name:      workload.GetName(),
					Namespace: workload.GetNamespace(),
				})
			}
		}

		for _, service := range servicesInMesh {
			serviceClusterName := serverMeshMetadata.ClusterName

			service.Spec.Federation = &smh_discovery_types.MeshServiceSpec_Federation{
				MulticlusterDnsName:  dns.BuildMulticlusterDnsName(service.Spec.GetKubeService().GetRef(), serviceClusterName),
				FederatedToWorkloads: federatedToWorkloads,
			}
		}

		err := updateServices(ctx, servicesInMesh, p.meshServiceClient)
		if err != nil {
			return err
		}
	}

	return nil
}
