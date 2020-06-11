package selection

import (
	"context"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

// these can technically be static functions
type BaseResourceSelector interface {
	FilterMeshServicesByServiceSelector(
		meshServices []*smh_discovery.MeshService,
		selector *smh_core_types.ServiceSelector,
	) ([]*smh_discovery.MeshService, error)

	FindMeshServiceByRefSelector(
		meshServices []*smh_discovery.MeshService,
		kubeServiceName string,
		kubeServiceNamespace string,
		kubeServiceCluster string,
	) *smh_discovery.MeshService
}

// Find Service Mesh Hub resources that correspond to k8s-native resources.
// Methods that do not receive a list of resources to filter or to select from will do a client lookup for those resources.
type ResourceSelector interface {
	BaseResourceSelector

	// fetch all MeshServices that match the given selector
	GetAllMeshServicesByServiceSelector(
		ctx context.Context,
		selector *smh_core_types.ServiceSelector,
	) ([]*smh_discovery.MeshService, error)

	// get the workloads that the given IdentitySelector applies to
	GetMeshWorkloadsByIdentitySelector(
		ctx context.Context,
		identitySelector *smh_core_types.IdentitySelector,
	) ([]*smh_discovery.MeshWorkload, error)

	// get the workloads that the given WorkloadSelector applies to
	GetMeshWorkloadsByWorkloadSelector(
		ctx context.Context,
		workloadSelector *smh_core_types.WorkloadSelector,
	) ([]*smh_discovery.MeshWorkload, error)

	// fetch the MeshService backing a k8s Service by the Service's name, namespace, cluster name
	// return error if no MeshService found, or multiple
	GetAllMeshServiceByRefSelector(
		ctx context.Context,
		kubeServiceName string,
		kubeServiceNamespace string,
		kubeServiceCluster string,
	) (*smh_discovery.MeshService, error)

	// get the Mesh Workload corresponding to the indicated pod controller (eg deployment)
	GetMeshWorkloadByRefSelector(
		ctx context.Context,
		podEventWatcherName string,
		podEventWatcherNamespace string,
		podEventWatcherCluster string,
	) (*smh_discovery.MeshWorkload, error)
}
