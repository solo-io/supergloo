package mesh

import (
	"context"

	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	rpc_v1 "github.com/solo-io/service-mesh-hub/services/apiserver/pkg/api/v1"
)

func NewMeshHandler(
	meshClient zephyr_discovery.MeshClient,
) rpc_v1.MeshApiServer {
	return &meshHandler{
		meshClient: meshClient,
	}
}

type meshHandler struct {
	meshClient zephyr_discovery.MeshClient
}

func (k *meshHandler) ListMeshes(
	ctx context.Context,
	_ *rpc_v1.ListMeshesRequest,
) (*rpc_v1.ListMeshesResponse, error) {
	clusters, err := k.meshClient.List(ctx)
	if err != nil {
		return nil, err
	}
	return &rpc_v1.ListMeshesResponse{
		Meshes: BuildRpcMeshList(clusters),
	}, nil
}

func BuildRpcMeshList(meshes *discovery_v1alpha1.MeshList) []*rpc_v1.Mesh {
	result := make([]*rpc_v1.Mesh, 0, len(meshes.Items))
	for _, v := range meshes.Items {
		result = append(result, BuildRpcMesh(&v))
	}
	return result
}

func BuildRpcMesh(mesh *discovery_v1alpha1.Mesh) *rpc_v1.Mesh {
	return &rpc_v1.Mesh{
		Spec: &mesh.Spec,
		Ref: &core_types.ResourceRef{
			Name:      mesh.GetName(),
			Namespace: mesh.GetNamespace(),
		},
		Labels: mesh.Labels,
	}
}
