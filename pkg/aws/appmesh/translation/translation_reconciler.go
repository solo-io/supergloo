package translation

import (
	"context"

	aws2 "github.com/aws/aws-sdk-go/aws"
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/collections/sets"
)

type appmeshTranslationReconciler struct {
	appmeshTranslator AppmeshTranslator
	dao               AppmeshAccessControlDao
}

func NewAppmeshTranslationReconciler(
	appmeshTranslator AppmeshTranslator,
	dao AppmeshAccessControlDao,
) AppmeshTranslationReconciler {
	return &appmeshTranslationReconciler{
		appmeshTranslator: appmeshTranslator,
		dao:               dao,
	}
}

func (a *appmeshTranslationReconciler) Reconcile(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
	virtualMesh *zephyr_networking.VirtualMesh,
) error {
	switch virtualMesh.Spec.GetEnforceAccessControl() {
	case zephyr_networking_types.VirtualMeshSpec_ENABLED, zephyr_networking_types.VirtualMeshSpec_MESH_DEFAULT:
		return a.reconcileWithEnforcedAccessControl(ctx, mesh)
	case zephyr_networking_types.VirtualMeshSpec_DISABLED:
		return a.reconcileWithDisabledAccessControl(ctx, mesh)
	}
	return nil
}

/*
	For every services declared as a target in at least one AccessControlPolicy, create an Appmesh VirtualService
	that routes to the VirtualNodes for the services' backing workloads.

	For every workload declared as a source in at least one AccessControlPolicy, create an Appmesh VirtualNode
	with VirtualServices corresponding to ACP declared upstream services as backends.
*/
func (a *appmeshTranslationReconciler) reconcileWithEnforcedAccessControl(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
) error {
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	servicesToBackingWorkloads, workloadsToBackingServices, err := a.dao.GetAllServiceWorkloadPairsForMesh(ctx, mesh)
	if err != nil {
		return err
	}
	servicesWithACP, err := a.dao.GetServicesWithACP(ctx, mesh)
	if err != nil {
		return err
	}
	workloadsWithACP, workloadsToUpstreamServices, err := a.dao.GetWorkloadsToUpstreamServicesWithACP(ctx, mesh)
	if err != nil {
		return err
	}
	for service, _ := range servicesToBackingWorkloads {
		if !servicesWithACP.Has(service) {
			delete(servicesToBackingWorkloads, service)
		}
	}
	for workload, _ := range workloadsToBackingServices {
		if !workloadsWithACP.Has(workload) {
			delete(workloadsToBackingServices, workload)
		}
	}
	// Create a route to allServices, and to upstream services
	err = a.reconcile(ctx, mesh, servicesToBackingWorkloads, workloadsToBackingServices, workloadsToUpstreamServices)
	return nil
}

/*
	For every services declared as a target in at least one AccessControlPolicy, create an Appmesh VirtualService
	that routes to the VirtualNodes for the services' backing workloads.

	For every workload declared as a source in at least one AccessControlPolicy, create an Appmesh VirtualNode
	with VirtualServices corresponding to all other upstream services in the Mesh as backends.
*/
func (a *appmeshTranslationReconciler) reconcileWithDisabledAccessControl(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
) error {
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	servicesToBackingWorkloads, workloadsToBackingServices, err := a.dao.GetAllServiceWorkloadPairsForMesh(ctx, mesh)
	if err != nil {
		return err
	}
	workloadsToAllUpstreamServices, err := a.dao.GetWorkloadsToAllUpstreamServices(ctx, mesh)
	if err != nil {
		return err
	}
	// Create a route to allServices, and to upstream services
	err = a.reconcile(ctx, mesh, servicesToBackingWorkloads, workloadsToBackingServices, workloadsToAllUpstreamServices)
	return nil
}

func (a *appmeshTranslationReconciler) reconcile(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
	servicesToBackingWorkloads map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload,
	workloadsToBackingServices map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService,
	workloadsToUpstreamServices map[*zephyr_discovery.MeshWorkload]sets.MeshServiceSet,
) error {
	logger := contextutils.LoggerFrom(ctx)
	var virtualServices []*appmesh2.VirtualServiceData
	var virtualRouters []*appmesh2.VirtualRouterData
	var routes []*appmesh2.RouteData
	var virtualNodes []*appmesh2.VirtualNodeData

	appmeshName := aws2.String(mesh.Spec.GetAwsAppMesh().GetName())
	for service, workloads := range servicesToBackingWorkloads {
		virtualService := a.appmeshTranslator.BuildVirtualService(appmeshName, service)
		virtualRouter := a.appmeshTranslator.BuildVirtualRouter(appmeshName, service)
		virtualServices = append(virtualServices, virtualService)
		virtualRouters = append(virtualRouters, virtualRouter)
		route, err := a.appmeshTranslator.BuildDefaultRoute(appmeshName, service, workloads)
		if err != nil {
			return err
		}
		routes = append(routes, route)
	}
	for workload, services := range workloadsToBackingServices {
		upstreamServices := workloadsToUpstreamServices[workload]
		var dnsService *zephyr_discovery.MeshService
		// For workloads represented by more than one k8s service, simply select the first service for DNS resolution.
		if len(services) > 0 {
			dnsService = services[0]
		}
		defaultVirtualNode := a.appmeshTranslator.BuildVirtualNode(appmeshName, workload, dnsService, upstreamServices.List())
		virtualNodes = append(virtualNodes, defaultVirtualNode)
	}
	err := a.dao.ReconcileVirtualNodes(ctx, mesh, virtualNodes)
	if err != nil {
		logger.Warnf("%+v", err)
	}
	err = a.dao.ReconcileVirtualRouters(ctx, mesh, virtualRouters)
	if err != nil {
		logger.Warnf("%+v", err)
	}
	err = a.dao.ReconcileVirtualServices(ctx, mesh, virtualServices)
	if err != nil {
		logger.Warnf("%+v", err)
	}
	err = a.dao.ReconcileRoutes(ctx, mesh, routes)
	if err != nil {
		logger.Warnf("%+v", err)
	}
	return nil
}
