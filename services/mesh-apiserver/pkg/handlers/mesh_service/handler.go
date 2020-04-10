package mesh_service

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	rpc_v1 "github.com/solo-io/service-mesh-hub/services/mesh-apiserver/pkg/api/v1"
)

func NewMeshServiceHandler(
	meshServiceClient zephyr_discovery.MeshServiceClient,
) rpc_v1.MeshServiceApiServer {
	return &meshServiceApiServer{
		meshServiceClient: meshServiceClient,
	}
}

type meshServiceApiServer struct {
	meshServiceClient zephyr_discovery.MeshServiceClient
}

func (k *meshServiceApiServer) ListMeshServices(
	ctx context.Context,
	_ *rpc_v1.ListMeshServicesRequest,
) (*rpc_v1.ListMeshServicesResponse, error) {
	clusters, err := k.meshServiceClient.List(ctx)
	if err != nil {
		return nil, err
	}
	return &rpc_v1.ListMeshServicesResponse{
		MeshServices: BuildRpcMeshServiceList(clusters),
	}, nil
}

func BuildRpcMeshServiceList(meshServices *discovery_v1alpha1.MeshServiceList) []*rpc_v1.MeshService {
	result := make([]*rpc_v1.MeshService, 0, len(meshServices.Items))
	for _, v := range meshServices.Items {
		result = append(result, BuildRpcMeshService(&v))
	}
	return result
}

func BuildRpcMeshService(meshService *discovery_v1alpha1.MeshService) *rpc_v1.MeshService {
	return &rpc_v1.MeshService{
		Spec: &meshService.Spec,
		Ref: &core_types.ResourceRef{
			Name:      meshService.GetName(),
			Namespace: meshService.GetNamespace(),
		},
		Labels: meshService.Labels,
	}
}
