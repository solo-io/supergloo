package kube

import "github.com/solo-io/service-mesh-hub/pkg/constants"

const (
	DISCOVERED_BY             = "discovered-by"
	MESH_WORKLOAD_DISCOVERY   = "mesh-workload-discovery"
	COMPUTE_TARGET            = "compute-target"
	KUBE_SERVICE_NAME         = "kube-service-name"
	KUBE_SERVICE_NAMESPACE    = "kube-service-namespace"
	KUBE_CONTROLLER_NAME      = "kube-controller-name"
	KUBE_CONTROLLER_NAMESPACE = "kube-controller-namespace"
	MESH_TYPE                 = "mesh-type"
)

var (
	OwnedBySMHLabel = map[string]string{"solo.io/owned-by": constants.ServiceMeshHubApplicationName}
)
