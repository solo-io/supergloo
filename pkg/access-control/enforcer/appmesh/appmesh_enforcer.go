package appmesh

import (
	"context"

	aws2 "github.com/aws/aws-sdk-go/aws"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/aws/appmesh"
)

const (
	EnforcerId = "appmesh_enforcer"
)

type appmeshEnforcer struct {
	appmeshTranslator appmesh.AppmeshTranslator
	dao               AppmeshAccessControlDao
}

func (a *appmeshEnforcer) Name() string {
	return EnforcerId
}

// TODO don't delete resources for network edges declared through AccessControlPolicies
// Delete VirtualServices and VirtualNodes unless explicitly declared in AccessControlPolicies.
func (a *appmeshEnforcer) StartEnforcing(ctx context.Context, mesh *zephyr_discovery.Mesh) error {
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	return nil
}

/*
	Ensure an Appmesh VirtualService for all k8s Services
	For every (workload, service) pair, declare the corresponding VirtualService (representing the k8s service)
	as a backend for the VirtualNode (representing the workload).
*/
func (a *appmeshEnforcer) StopEnforcing(ctx context.Context, mesh *zephyr_discovery.Mesh) error {
	if mesh.Spec.GetAwsAppMesh() == nil {
		return nil
	}
	serviceToWorkloads, workloadToServices, err := a.dao.GetServicesAndWorkloadsForMesh(ctx, mesh)
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
		err = a.dao.EnsureVirtualService(virtualService)
		if err != nil {
			return err
		}
		err = a.dao.EnsureVirtualRouter(virtualRouter)
		if err != nil {
			return err
		}
		err = a.dao.EnsureRoute(defaultRoute)
		if err != nil {
			return err
		}
	}
	for workload, services := range workloadToServices {
		var service *zephyr_discovery.MeshService
		if len(services) > 0 {
			service = services[0]
		}
		defaultVirtualNode := a.appmeshTranslator.BuildDefaultVirtualNode(appmeshName, workload, service, services)
		err = a.dao.EnsureVirtualNode(defaultVirtualNode)
		if err != nil {
			return err
		}
	}
	return nil
}
