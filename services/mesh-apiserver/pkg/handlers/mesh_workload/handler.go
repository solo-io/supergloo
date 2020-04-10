package mesh_workload

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	rpc_v1 "github.com/solo-io/service-mesh-hub/services/mesh-apiserver/pkg/api/v1"
)

func NewMeshWorkloadHandler(
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
) rpc_v1.MeshWorkloadApiServer {
	return &meshWorkloadApiServer{
		meshWorkloadClient: meshWorkloadClient,
	}
}

type meshWorkloadApiServer struct {
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient
}

func (k *meshWorkloadApiServer) ListMeshWorkloads(
	ctx context.Context,
	_ *rpc_v1.ListMeshWorkloadsRequest,
) (*rpc_v1.ListMeshWorkloadsResponse, error) {
	clusters, err := k.meshWorkloadClient.List(ctx)
	if err != nil {
		return nil, err
	}
	return &rpc_v1.ListMeshWorkloadsResponse{
		MeshWorkloads: BuildRpcMeshWorkloadList(clusters),
	}, nil
}

func BuildRpcMeshWorkloadList(meshWorkloads *discovery_v1alpha1.MeshWorkloadList) []*rpc_v1.MeshWorkload {
	result := make([]*rpc_v1.MeshWorkload, 0, len(meshWorkloads.Items))
	for _, v := range meshWorkloads.Items {
		result = append(result, BuildRpcMeshWorkload(&v))
	}
	return result
}

func BuildRpcMeshWorkload(meshWorkload *discovery_v1alpha1.MeshWorkload) *rpc_v1.MeshWorkload {
	return &rpc_v1.MeshWorkload{
		Spec: &meshWorkload.Spec,
		Ref: &core_types.ResourceRef{
			Name:      meshWorkload.GetName(),
			Namespace: meshWorkload.GetNamespace(),
		},
		Labels: meshWorkload.Labels,
	}
}
