package handlers

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/handlers/kubernetes_cluster"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/handlers/mesh"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/handlers/mesh_service"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/handlers/mesh_workload"
	"github.com/solo-io/service-mesh-hub/services/apiserver/pkg/handlers/virtual_mesh"
)

var HandlerSet = wire.NewSet(
	kubernetes_cluster.NewKubernetesClusterHandler,
	mesh.NewMeshHandler,
	mesh_service.NewMeshServiceHandler,
	mesh_workload.NewMeshWorkloadHandler,
	virtual_mesh.NewVirtualMeshHandler,
)
