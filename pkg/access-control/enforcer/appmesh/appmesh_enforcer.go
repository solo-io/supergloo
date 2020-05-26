package appmesh

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
)

const (
	EnforcerId = "appmesh_enforcer"
)

type appmeshEnforcer struct {
	dao AppmeshAccessControlDao
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
	meshServices, err := a.dao.ListMeshServicesForMesh(ctx, mesh)
	if err != nil {
		return err
	}
	var virtualServiceNames []string
	for _, meshService := range meshServices {
		virtualServiceRef, err := a.dao.EnsureAppmeshVirtualService(ctx, mesh, meshService)
		if err != nil {
			return err
		}
		virtualServiceNames = append(virtualServiceNames, aws.StringValue(virtualServiceRef.VirtualServiceName))
	}
	meshWorkloads, err := a.dao.ListMeshWorkloadsForMesh(ctx, mesh)
	if err != nil {
		return err
	}
	for _, meshWorkload := range meshWorkloads {
		err = a.dao.EnsureVirtualNodeBackends(ctx, mesh, meshWorkload, virtualServiceNames)
		if err != nil {
			return err
		}
	}
	return nil
}
