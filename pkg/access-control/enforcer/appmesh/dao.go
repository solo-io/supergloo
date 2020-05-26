package appmesh

import (
	"context"
	"strings"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/aws"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/clients/appmesh"
)

type appmeshAccessControlDao struct {
	meshServiceClient    zephyr_discovery.MeshServiceClient
	meshWorkloadClient   zephyr_discovery.MeshWorkloadClient
	appmeshClientFactory appmesh.AppMeshClientFactory
	awsCredentialsGetter aws.AwsCredentialsGetter
}

func (a *appmeshAccessControlDao) ListMeshServicesForMesh(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
) ([]*zephyr_discovery.MeshService, error) {
	meshServiceList, err := a.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return nil, err
	}
	var meshServices []*zephyr_discovery.MeshService
	for _, meshService := range meshServiceList.Items {
		meshService := meshService
		if meshService.Spec.GetMesh().GetName() != mesh.GetName() || meshService.Spec.GetMesh().GetNamespace() != mesh.GetNamespace() {
			continue
		}
		meshServices = append(meshServices, &meshService)
	}
	return meshServices, nil
}

func (a *appmeshAccessControlDao) ListMeshWorkloadsForMesh(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
) ([]*zephyr_discovery.MeshWorkload, error) {
	meshWorkloadList, err := a.meshWorkloadClient.ListMeshWorkload(ctx)
	if err != nil {
		return nil, err
	}
	var meshWorkloads []*zephyr_discovery.MeshWorkload
	for _, meshWorkload := range meshWorkloadList.Items {
		meshWorkload := meshWorkload
		if meshWorkload.Spec.GetMesh().GetName() != mesh.GetName() || meshWorkload.Spec.GetMesh().GetNamespace() != mesh.GetNamespace() {
			continue
		}
		meshWorkloads = append(meshWorkloads, &meshWorkload)
	}
	return meshWorkloads, nil
}

func (a *appmeshAccessControlDao) EnsureAppmeshVirtualServiceWithRouter(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
	meshService *zephyr_discovery.MeshService,
) error {
	appmeshClient, err := a.buildAppmeshClient(mesh)
	if err != nil {
		return nil
	}
	meshName := aws2.String(mesh.Spec.GetAwsAppMesh().GetName())
	virtualServiceName := aws2.String(metadata.BuildVirtualServiceName(meshService))
	virtualRouterName := aws2.String(metadata.BuildVirtualRouterName(meshService))
	virtualNodesName := a.resourceSelector.GetBackingMeshWorkloadsForService(ctx, meshService)
	a.ensureVirtualServiceWithRouterProvider(appmeshClient, meshName, virtualServiceName, virtualRouterName)
	a.ensureVirtualRouter(appmeshClient, meshName, virtualRouterName, meshService.Spec.GetKubeService().GetPorts())
	a.ensureRouteToAllWorkloads(appmeshClient, meshName, virtualRouterName, virtualNodeNames)
	return err
}

func (a *appmeshAccessControlDao) EnsureVirtualNodeBackends(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
	meshWorkload *zephyr_discovery.MeshWorkload,
	virtualServiceNames []string,
) error {
	appmeshClient, err := a.buildAppmeshClient(mesh)
	if err != nil {
		return nil
	}
	return nil
}

func (a *appmeshAccessControlDao) buildAppmeshClient(mesh *zephyr_discovery.Mesh) (appmeshiface.AppMeshAPI, error) {
	creds, err := a.awsCredentialsGetter.Get(mesh.Spec.GetAwsAppMesh().GetAwsAccountId())
	if err != nil {
		return nil, err
	}
	appmeshClient, err := a.appmeshClientFactory(creds, mesh.Spec.GetAwsAppMesh().GetRegion())
	if err != nil {
		return nil, err
	}
	return appmeshClient, nil
}

func (a *appmeshAccessControlDao) ensureVirtualServiceWithRouterProvider(
	appmeshClient appmeshiface.AppMeshAPI,
	meshName *string,
	virtualServiceName *string,
	virtualRouterName *string,
) error {
	virtualService, err := appmeshClient.DescribeVirtualService(
		&appmesh2.DescribeVirtualServiceInput{
			MeshName:           meshName,
			VirtualServiceName: virtualServiceName,
		},
	)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case appmesh2.ErrCodeNotFoundException:
				break
			default:
				return err
			}
		}
	}
	if virtualService.VirtualService.Spec.Provider.VirtualRouter != nil {
		return nil
	}
	_, err = appmeshClient.CreateVirtualService(
		&appmesh2.CreateVirtualServiceInput{
			MeshName: meshName,
			Spec: &appmesh2.VirtualServiceSpec{
				Provider: &appmesh2.VirtualServiceProvider{
					VirtualRouter: &appmesh2.VirtualRouterServiceProvider{
						VirtualRouterName: virtualRouterName,
					},
				},
			},
			VirtualServiceName: virtualServiceName,
		},
	)
	return err
}

// Ensure that a VirtualRouter exists with all k8s Service ports
func (a *appmeshAccessControlDao) ensureVirtualRouter(
	appmeshClient appmeshiface.AppMeshAPI,
	meshName *string,
	virtualRouterName *string,
	ports []*types.MeshServiceSpec_KubeService_KubeServicePort,
) error {
	virtualRouter, err := appmeshClient.DescribeVirtualRouter(
		&appmesh2.DescribeVirtualRouterInput{
			MeshName:          meshName,
			VirtualRouterName: virtualRouterName,
		})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case appmesh2.ErrCodeNotFoundException:
				break
			default:
				return err
			}
		}
	}
	servicePorts := map[int64]string{}
	for _, port := range ports {
		servicePorts[int64(port.GetPort())] = port.GetProtocol()
	}
	shouldCreate := false
	for _, listener := range virtualRouter.VirtualRouter.Spec.Listeners {
		listenerPort := aws2.Int64Value(listener.PortMapping.Port)
		protocol, ok := servicePorts[listenerPort]
		if !ok || strings.ToLower(protocol) != strings.ToLower(aws2.StringValue(listener.PortMapping.Protocol)) {
			shouldCreate = true
			break
		}
		delete(servicePorts, listenerPort)
	}
	if !shouldCreate && len(servicePorts) == 0 {
		return nil
	}
	var virtualRouterListeners []*appmesh2.VirtualRouterListener
	for _, servicePort := range ports {
		virtualRouterListeners = append(virtualRouterListeners, &appmesh2.VirtualRouterListener{
			PortMapping: &appmesh2.PortMapping{
				Port:     aws2.Int64(int64(servicePort.GetPort())),
				Protocol: aws2.String(servicePort.GetProtocol()),
			},
		})
	}
	_, err = appmeshClient.CreateVirtualRouter(
		&appmesh2.CreateVirtualRouterInput{
			MeshName: meshName,
			Spec: &appmesh2.VirtualRouterSpec{
				Listeners: virtualRouterListeners,
			},
			VirtualRouterName: virtualRouterName,
		},
	)
	return err
}

// For a VirtualRouter, ensure a route that targets all VirtualNodes
// (representing the workloads) backing the k8s service with equal weights.
func (a *appmeshAccessControlDao) ensureRouteToAllWorkloads(
	appmeshClient appmeshiface.AppMeshAPI,
	meshName *string,
	virtualRouterName *string,
	virtualNodeNames []*string,
) error {

}
