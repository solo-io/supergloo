package translation

import (
	"context"

	aws2 "github.com/aws/aws-sdk-go/aws"
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/go-multierror"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
)

const (
	// Canonical name for default route that permits traffic to all workloads backing service with equal weight.
	DefaultRouteName = "smh-default"
	// Default route always takes lowest priority (ranges from 0-1000 inclusive).
	DefaultRoutePriority = 1000
)

type appmeshTranslationReconciler struct {
	appmeshTranslator AppmeshTranslator
	dao               AppmeshTranslationDao
}

func NewAppmeshTranslationReconciler(
	appmeshTranslator AppmeshTranslator,
	dao AppmeshTranslationDao,
) AppmeshTranslationReconciler {
	return &appmeshTranslationReconciler{
		appmeshTranslator: appmeshTranslator,
		dao:               dao,
	}
}

// If vm is populated, then output Appmesh resources to enable all workload/service communication.
// If vm is nil, then ensure that no Appmesh resources exist.
// See https://github.com/solo-io/service-mesh-hub/issues/750
func (a *appmeshTranslationReconciler) Reconcile(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
	vm *smh_networking.VirtualMesh,
) error {
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	if vm == nil {
		// Delete all translated Appmesh resources.
		return a.reconcile(ctx, mesh, nil, nil, nil)
	}
	// Be default, configure Appmesh envoy sidecars to allow traffic from any workload to any service.
	return a.reconcileWithDisabledAccessControl(ctx, mesh)
}

/*
	TODO: wire up this method when SMH API exposes sidecar configuration options
	For every services declared as a target in at least one AccessControlPolicy, create an Appmesh VirtualService
	that routes to the VirtualNodes for the services' backing workloads.

	For every workload declared as a source in at least one AccessControlPolicy, create an Appmesh VirtualNode
	with VirtualServices corresponding to ACP declared upstream services as backends.
*/
func (a *appmeshTranslationReconciler) reconcileWithEnforcedAccessControl(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
) error {
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
	// Filter out services/workloads that are not declared in ACPs
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
	return err
}

/*
	Configure Appmesh envoy sidecars to allow traffic between any (workload, service) pair.
*/
func (a *appmeshTranslationReconciler) reconcileWithDisabledAccessControl(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
) error {
	servicesToBackingWorkloads, workloadsToBackingServices, err := a.dao.GetAllServiceWorkloadPairsForMesh(ctx, mesh)
	if err != nil {
		return err
	}
	workloadsToAllUpstreamServices, err := a.dao.GetWorkloadsToAllUpstreamServices(ctx, mesh)
	if err != nil {
		return err
	}
	// Create a route to allServices, and to upstream services
	return a.reconcile(ctx, mesh, servicesToBackingWorkloads, workloadsToBackingServices, workloadsToAllUpstreamServices)
}

func (a *appmeshTranslationReconciler) reconcile(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
	servicesToBackingWorkloads map[*smh_discovery.MeshService][]*smh_discovery.MeshWorkload,
	workloadsToBackingServices map[*smh_discovery.MeshWorkload][]*smh_discovery.MeshService,
	workloadsToUpstreamServices map[string]smh_discovery_sets.MeshServiceSet,
) error {
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
		// Build default Route that routes to all backing workloads with equal weight.
		route, err := a.appmeshTranslator.BuildRoute(appmeshName, DefaultRouteName, DefaultRoutePriority, service, workloads)
		if err != nil {
			return err
		}
		routes = append(routes, route)
	}
	for workload, services := range workloadsToBackingServices {
		// TODO: add Cloudmap support, https://github.com/solo-io/service-mesh-hub/issues/755
		// Don't create VirtualNode for workloads not backed by service because there's no DNS resolution.
		if len(services) == 0 {
			continue
		}
		var upstreamServicesList []*smh_discovery.MeshService
		upstreamServices := workloadsToUpstreamServices[selection.ToUniqueSingleClusterString(workload.ObjectMeta)]
		if upstreamServices != nil {
			upstreamServicesList = upstreamServices.List()
		}
		// For workloads represented by more than one k8s service, simply select the first service for DNS resolution.
		dnsService := services[0]
		defaultVirtualNode := a.appmeshTranslator.BuildVirtualNode(appmeshName, workload, dnsService, upstreamServicesList)
		virtualNodes = append(virtualNodes, defaultVirtualNode)
	}
	var multierr *multierror.Error
	err := a.dao.ReconcileVirtualNodes(ctx, mesh, virtualNodes)
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}
	err = a.dao.ReconcileVirtualRoutersAndRoutesAndVirtualServices(ctx, mesh, virtualRouters, routes, virtualServices)
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}
	return multierr.ErrorOrNil()
}
