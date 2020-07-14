package translation

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AppmeshTranslationReconciler interface {
	Reconcile(ctx context.Context,
		mesh *smh_discovery.Mesh,
		virtualMesh *smh_networking.VirtualMesh,
	) error
}

type AppmeshTranslator interface {
	// For a given MeshWorkload, return a VirtualNode with the given upstream services declared as VirtualService backends.
	BuildVirtualNode(
		appmeshName *string,
		meshWorkload *smh_discovery.MeshWorkload,
		meshService *smh_discovery.MeshService,
		upstreamServices []*smh_discovery.MeshService,
	) *appmesh.VirtualNodeData

	// For a given MeshService, return a Route that targets all given workloads with equal weight.
	BuildRoute(
		appmeshName *string,
		routeName string,
		priority int,
		meshService *smh_discovery.MeshService,
		targetedWorkloads []*smh_discovery.MeshWorkload,
	) (*appmesh.RouteData, error)

	// For a given MeshService, return a VirtualService.
	BuildVirtualService(
		appmeshName *string,
		meshService *smh_discovery.MeshService,
	) *appmesh.VirtualServiceData

	// For a given MeshService, return a VirtualRouter.
	BuildVirtualRouter(
		appmeshName *string,
		meshService *smh_discovery.MeshService,
	) *appmesh.VirtualRouterData
}

type AppmeshTranslationDao interface {
	// Return two maps which associate workloads to backing services and vice versa.
	GetAllServiceWorkloadPairsForMesh(
		ctx context.Context,
		mesh *smh_discovery.Mesh,
	) (map[*smh_discovery.MeshService][]*smh_discovery.MeshWorkload,
		map[*smh_discovery.MeshWorkload][]*smh_discovery.MeshService,
		error)

	GetWorkloadsToAllUpstreamServices(
		ctx context.Context,
		mesh *smh_discovery.Mesh,
	) (map[string]smh_discovery_sets.MeshServiceSet, error)

	GetServicesWithACP(
		ctx context.Context,
		mesh *smh_discovery.Mesh,
	) (smh_discovery_sets.MeshServiceSet, error)

	GetWorkloadsToUpstreamServicesWithACP(
		ctx context.Context,
		mesh *smh_discovery.Mesh,
	) (smh_discovery_sets.MeshWorkloadSet, map[string]smh_discovery_sets.MeshServiceSet, error)

	// These need to be reconciled as a unit because of the ordering constraints imposed by the AWS API.
	ReconcileVirtualRoutersAndRoutesAndVirtualServices(
		ctx context.Context,
		mesh *smh_discovery.Mesh,
		virtualRouters []*appmesh.VirtualRouterData,
		routes []*appmesh.RouteData,
		virtualServices []*appmesh.VirtualServiceData,
	) error

	ReconcileVirtualNodes(
		ctx context.Context,
		mesh *smh_discovery.Mesh,
		virtualNodes []*appmesh.VirtualNodeData,
	) error
}
