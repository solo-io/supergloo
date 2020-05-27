package appmesh

import (
	"context"

	"github.com/aws/aws-sdk-go/service/appmesh"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	appmesh2 "github.com/solo-io/service-mesh-hub/pkg/aws/appmesh"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	"k8s.io/apimachinery/pkg/labels"
)

type appmeshAccessControlDao struct {
	meshServiceClient    zephyr_discovery.MeshServiceClient
	meshWorkloadClient   zephyr_discovery.MeshWorkloadClient
	resourceSelector     selector.ResourceSelector
	appmeshClientFactory appmesh2.AppmeshClientFactory
}

func NewAppmeshAccessControlDao(
	meshServiceClient zephyr_discovery.MeshServiceClient,
	meshWorkloadClient zephyr_discovery.MeshWorkloadClient,
	resourceSelector selector.ResourceSelector,
	appmeshClientFactory appmesh2.AppmeshClientFactory,
) AppmeshAccessControlDao {
	return &appmeshAccessControlDao{
		meshServiceClient:    meshServiceClient,
		meshWorkloadClient:   meshWorkloadClient,
		resourceSelector:     resourceSelector,
		appmeshClientFactory: appmeshClientFactory,
	}
}

func (a *appmeshAccessControlDao) GetServicesAndWorkloadsForMesh(
	ctx context.Context,
	mesh *zephyr_discovery.Mesh,
) (map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload,
	map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService,
	error) {
	meshServices, err := a.listMeshServicesForMesh(ctx, mesh)
	if err != nil {
		return nil, nil, err
	}
	meshWorkloads, err := a.listMeshWorkloadsForMesh(ctx, mesh)
	if err != nil {
		return nil, nil, err
	}
	serviceToWorkloads := map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload{}
	workloadToServices := map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService{}
	for _, workload := range meshWorkloads {
		workload := workload
		for _, service := range meshServices {
			if isServiceBackedByWorkload(service, workload) {
				serviceToWorkloads[service] = append(serviceToWorkloads[service], workload)
				workloadToServices[workload] = append(workloadToServices[workload], service)
			}
		}
	}
	for _, service := range meshServices {
		_, ok := serviceToWorkloads[service]
		if !ok {
			serviceToWorkloads[service] = nil
		}
	}
	return serviceToWorkloads, workloadToServices, nil
}

func (a *appmeshAccessControlDao) listMeshServicesForMesh(
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

func (a *appmeshAccessControlDao) listMeshWorkloadsForMesh(
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

func (a *appmeshAccessControlDao) EnsureVirtualService(
	mesh *zephyr_discovery.Mesh,
	virtualServiceData *appmesh.VirtualServiceData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	return appmeshClient.EnsureVirtualService(virtualServiceData)
}
func (a *appmeshAccessControlDao) EnsureVirtualRouter(
	mesh *zephyr_discovery.Mesh,
	virtualRouter *appmesh.VirtualRouterData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	return appmeshClient.EnsureVirtualRouter(virtualRouter)
}
func (a *appmeshAccessControlDao) EnsureRoute(
	mesh *zephyr_discovery.Mesh,
	route *appmesh.RouteData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	return appmeshClient.EnsureRoute(route)
}
func (a *appmeshAccessControlDao) EnsureVirtualNode(
	mesh *zephyr_discovery.Mesh,
	virtualNode *appmesh.VirtualNodeData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	return appmeshClient.EnsureVirtualNode(virtualNode)
}

func isServiceBackedByWorkload(
	meshService *zephyr_discovery.MeshService,
	meshWorkload *zephyr_discovery.MeshWorkload,
) bool {
	if len(meshService.Spec.GetKubeService().GetWorkloadSelectorLabels()) == 0 ||
		len(meshWorkload.Spec.GetKubeController().GetLabels()) == 0 {
		return false
	}
	return labels.AreLabelsInWhiteList(
		meshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
		meshWorkload.Spec.GetKubeController().GetLabels(),
	)
}
