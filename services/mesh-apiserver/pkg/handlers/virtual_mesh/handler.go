package virtual_mesh

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking"
	rpc_v1 "github.com/solo-io/service-mesh-hub/services/mesh-apiserver/pkg/api/v1"
)

func NewVirtualMeshHandler(
	meshWorkloadClient zephyr_networking.VirtualMeshClient,
) rpc_v1.VirtualMeshApiServer {
	return &meshWorkloadApiServer{
		virtualMeshClient: meshWorkloadClient,
	}
}

type meshWorkloadApiServer struct {
	virtualMeshClient zephyr_networking.VirtualMeshClient
}

func (k *meshWorkloadApiServer) ListVirtualMeshes(
	ctx context.Context,
	_ *rpc_v1.ListVirtualMeshesRequest,
) (*rpc_v1.ListVirtualMeshesResponse, error) {
	clusters, err := k.virtualMeshClient.List(ctx)
	if err != nil {
		return nil, err
	}
	return &rpc_v1.ListVirtualMeshesResponse{
		VirtualMeshes: BuildRpcVirtualMeshList(clusters),
	}, nil
}

func BuildRpcVirtualMeshList(virtualMeshList *networking_v1alpha1.VirtualMeshList) []*rpc_v1.VirtualMesh {
	result := make([]*rpc_v1.VirtualMesh, 0, len(virtualMeshList.Items))
	for _, v := range virtualMeshList.Items {
		result = append(result, BuildRpcVirtualMesh(&v))
	}
	return result
}

func BuildRpcVirtualMesh(virtualMesh *networking_v1alpha1.VirtualMesh) *rpc_v1.VirtualMesh {
	return &rpc_v1.VirtualMesh{
		Spec: &virtualMesh.Spec,
		Ref: &core_types.ResourceRef{
			Name:      virtualMesh.GetName(),
			Namespace: virtualMesh.GetNamespace(),
		},
		Labels: virtualMesh.Labels,
	}
}
