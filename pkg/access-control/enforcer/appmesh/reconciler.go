package appmesh

import (
	"context"

	aws2 "github.com/aws/aws-sdk-go/aws"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/aws/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/collections/sets"
)

type appmeshAccessControlReconciler struct {
	appmeshTranslator appmesh.AppmeshTranslator
	dao               AppmeshAccessControlDao
}

func (a *appmeshAccessControlReconciler) reconcile(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
	virtualMesh *zephyr_networking.VirtualMesh,
) error {
	switch virtualMesh.Spec.GetEnforceAccessControl() {
	case zephyr_networking_types.VirtualMeshSpec_ENABLED, zephyr_networking_types.VirtualMeshSpec_MESH_DEFAULT:
		return a.reconcileWithEnforcement(ctx, mesh)
	case zephyr_networking_types.VirtualMeshSpec_DISABLED:
		return a.reconcileWithoutEnforcement(ctx, mesh)
	}
	return nil
}

func (a *appmeshAccessControlReconciler) reconcileWithEnforcement(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
) error {
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	err := a.dao.DeleteAllDefaultRoutes(mesh)
	if err != nil {
		return err
	}
	serviceToWorkloads, workloadToServices, err := a.dao.GetServicesWorkloadPairsForMesh(ctx, mesh)
	if err != nil {
		return err
	}
	servicesWithACP, err := a.dao.GetServicesWithACP(ctx, mesh)
	if err != nil {
		return err
	}
	workloadsWithACP, err := a.dao.GetWorkloadsWithACP(ctx, mesh)
	if err != nil {
		return err
	}
	for service, _ := range serviceToWorkloads {
		if !servicesWithACP.Has(service) {
			delete(serviceToWorkloads, service)
		}
	}
	for workload, _ := range workloadToServices {
		if !workloadsWithACP.Has(workload) {
			delete(workloadToServices, workload)
		}
	}
	// Ensure Appmesh entity collection state matches declared here
	return nil
}

/*
	Ensure an Appmesh VirtualService for all k8s Services
	For every (workload, service) pair, declare the corresponding VirtualService (representing the k8s service)
	as a backend for the VirtualNode (representing the workload).
*/
func (a *appmeshAccessControlReconciler) reconcileWithoutEnforcement(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
) error {
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	serviceToWorkloads, workloadToServices, err := a.dao.GetServicesWorkloadPairsForMesh(ctx, mesh)
	if err != nil {
		return err
	}
	appmeshName := aws2.String(mesh.Spec.GetAwsAppMesh().GetName())
	for service, workloads := range serviceToWorkloads {
		virtualService := a.appmeshTranslator.BuildVirtualService(appmeshName, service)
		virtualRouter := a.appmeshTranslator.BuildVirtualRouter(appmeshName, service)
		defaultRoute, err := a.appmeshTranslator.BuildDefaultRoute(appmeshName, service, workloads)
		if err != nil {
			return err
		}
		err = a.dao.EnsureVirtualService(mesh, virtualService)
		if err != nil {
			return err
		}
		err = a.dao.EnsureVirtualRouter(mesh, virtualRouter)
		if err != nil {
			return err
		}
		err = a.dao.EnsureRoute(mesh, defaultRoute)
		if err != nil {
			return err
		}
	}
	allServices := sets.NewMeshServiceSet()
	for service, _ := range serviceToWorkloads {
		allServices.Insert(service)
	}
	for workload, services := range workloadToServices {
		servicesSet := sets.NewMeshServiceSet(services...)
		upstreamServices := allServices.Difference(servicesSet).List()
		var dnsService *zephyr_discovery.MeshService
		// For workloads represented by more than one k8s service, simply select the first service for DNS resolution.
		if len(services) > 0 {
			dnsService = services[0]
		}
		defaultVirtualNode := a.appmeshTranslator.BuildDefaultVirtualNode(appmeshName, workload, dnsService, upstreamServices)
		err = a.dao.EnsureVirtualNode(mesh, defaultVirtualNode)
		if err != nil {
			return err
		}
	}
	return nil
}
