package translation

import (
	"context"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	appmesh2 "github.com/solo-io/service-mesh-hub/pkg/aws/clients"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	"k8s.io/apimachinery/pkg/labels"
)

type appmeshTranslationDao struct {
	meshServiceClient    smh_discovery.MeshServiceClient
	meshWorkloadClient   smh_discovery.MeshWorkloadClient
	acpClient            smh_networking.AccessControlPolicyClient
	resourceSelector     selection.ResourceSelector
	appmeshClientFactory appmesh2.AppmeshClientGetter
}

func NewAppmeshAccessControlDao(
	meshServiceClient smh_discovery.MeshServiceClient,
	meshWorkloadClient smh_discovery.MeshWorkloadClient,
	resourceSelector selection.ResourceSelector,
	appmeshClientFactory appmesh2.AppmeshClientGetter,
	acpClient smh_networking.AccessControlPolicyClient,
) AppmeshTranslationDao {
	return &appmeshTranslationDao{
		meshServiceClient:    meshServiceClient,
		meshWorkloadClient:   meshWorkloadClient,
		resourceSelector:     resourceSelector,
		appmeshClientFactory: appmeshClientFactory,
		acpClient:            acpClient,
	}
}

func (a *appmeshTranslationDao) GetAllServiceWorkloadPairsForMesh(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
) (map[*smh_discovery.MeshService][]*smh_discovery.MeshWorkload,
	map[*smh_discovery.MeshWorkload][]*smh_discovery.MeshService,
	error) {
	meshServices, err := a.listMeshServicesForMesh(ctx, mesh)
	if err != nil {
		return nil, nil, err
	}
	meshWorkloads, err := a.listMeshWorkloadsForMesh(ctx, mesh)
	if err != nil {
		return nil, nil, err
	}
	serviceToWorkloads, workloadToServices := a.buildServiceToWorkloadMaps(meshServices, meshWorkloads)
	return serviceToWorkloads, workloadToServices, nil
}

func (a *appmeshTranslationDao) GetWorkloadsToAllUpstreamServices(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
) (map[string]smh_discovery_sets.MeshServiceSet, error) {
	meshServices, err := a.listMeshServicesForMesh(ctx, mesh)
	if err != nil {
		return nil, err
	}
	meshWorkloads, err := a.listMeshWorkloadsForMesh(ctx, mesh)
	if err != nil {
		return nil, err
	}
	meshServiceSet := smh_discovery_sets.NewMeshServiceSet(meshServices...)
	workloadsToAllUpstreamServices := map[string]smh_discovery_sets.MeshServiceSet{}
	for _, meshWorkload := range meshWorkloads {
		workloadsToAllUpstreamServices[selection.ToUniqueSingleClusterString(meshWorkload.ObjectMeta)] = meshServiceSet
	}
	return workloadsToAllUpstreamServices, nil
}

func (a *appmeshTranslationDao) GetServicesWithACP(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
) (smh_discovery_sets.MeshServiceSet, error) {
	meshServices, err := a.listMeshServicesForMesh(ctx, mesh)
	if err != nil {
		return nil, err
	}
	servicesInMesh := smh_discovery_sets.NewMeshServiceSet(meshServices...)
	acpList, err := a.acpClient.ListAccessControlPolicy(ctx)
	if err != nil {
		return nil, err
	}
	services := smh_discovery_sets.NewMeshServiceSet()
	for _, acp := range acpList.Items {
		acpServices, err := a.resourceSelector.GetAllMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector())
		if err != nil {
			return nil, err
		}
		acpServicesSet := smh_discovery_sets.NewMeshServiceSet(acpServices...)
		services.Insert(servicesInMesh.Intersection(acpServicesSet).List()...)
	}
	return services, nil
}

func (a *appmeshTranslationDao) GetWorkloadsToUpstreamServicesWithACP(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
) (smh_discovery_sets.MeshWorkloadSet, map[string]smh_discovery_sets.MeshServiceSet, error) {
	meshWorkloads, err := a.listMeshWorkloadsForMesh(ctx, mesh)
	if err != nil {
		return nil, nil, err
	}
	workloadsInMesh := smh_discovery_sets.NewMeshWorkloadSet(meshWorkloads...)
	workloadsToUpstreamServices := map[string]smh_discovery_sets.MeshServiceSet{}
	acpList, err := a.acpClient.ListAccessControlPolicy(ctx)
	if err != nil {
		return nil, nil, err
	}
	declaredWorkloads := smh_discovery_sets.NewMeshWorkloadSet()
	for _, acp := range acpList.Items {
		workloads, err := a.resourceSelector.GetMeshWorkloadsByIdentitySelector(ctx, acp.Spec.GetSourceSelector())
		upstreamServices, err := a.resourceSelector.GetAllMeshServicesByServiceSelector(ctx, acp.Spec.GetDestinationSelector())
		if err != nil {
			return nil, nil, err
		}
		for _, workload := range workloads {
			workloadKey := selection.ToUniqueSingleClusterString(workload.ObjectMeta)
			upstreamServicesSet, ok := workloadsToUpstreamServices[workloadKey]
			if !ok {
				workloadsToUpstreamServices[workloadKey] = smh_discovery_sets.NewMeshServiceSet(upstreamServices...)
			} else {
				upstreamServicesSet.Insert(upstreamServices...)
			}
		}
		acpWorkloadsSet := smh_discovery_sets.NewMeshWorkloadSet(workloads...)
		declaredWorkloads.Insert(workloadsInMesh.Intersection(acpWorkloadsSet).List()...)
	}
	return declaredWorkloads, workloadsToUpstreamServices, nil
}

func (a *appmeshTranslationDao) buildServiceToWorkloadMaps(
	meshServices []*smh_discovery.MeshService,
	meshWorkloads []*smh_discovery.MeshWorkload,
) (map[*smh_discovery.MeshService][]*smh_discovery.MeshWorkload,
	map[*smh_discovery.MeshWorkload][]*smh_discovery.MeshService) {
	serviceToWorkloads := map[*smh_discovery.MeshService][]*smh_discovery.MeshWorkload{}
	workloadToServices := map[*smh_discovery.MeshWorkload][]*smh_discovery.MeshService{}
	for _, workload := range meshWorkloads {
		workloadToServices[workload] = nil
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
	return serviceToWorkloads, workloadToServices
}

func (a *appmeshTranslationDao) listMeshServicesForMesh(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
) ([]*smh_discovery.MeshService, error) {
	meshServiceList, err := a.meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return nil, err
	}
	var meshServices []*smh_discovery.MeshService
	for _, meshService := range meshServiceList.Items {
		meshService := meshService
		if meshService.Spec.GetMesh().GetName() != mesh.GetName() || meshService.Spec.GetMesh().GetNamespace() != mesh.GetNamespace() {
			continue
		}
		meshServices = append(meshServices, &meshService)
	}
	return meshServices, nil
}

func (a *appmeshTranslationDao) listMeshWorkloadsForMesh(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
) ([]*smh_discovery.MeshWorkload, error) {
	meshWorkloadList, err := a.meshWorkloadClient.ListMeshWorkload(ctx)
	if err != nil {
		return nil, err
	}
	var meshWorkloads []*smh_discovery.MeshWorkload
	for _, meshWorkload := range meshWorkloadList.Items {
		meshWorkload := meshWorkload
		if meshWorkload.Spec.GetMesh().GetName() != mesh.GetName() || meshWorkload.Spec.GetMesh().GetNamespace() != mesh.GetNamespace() {
			continue
		}
		meshWorkloads = append(meshWorkloads, &meshWorkload)
	}
	return meshWorkloads, nil
}

func (a *appmeshTranslationDao) EnsureVirtualService(
	mesh *smh_discovery.Mesh,
	virtualServiceData *appmesh.VirtualServiceData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	return appmeshClient.EnsureVirtualService(virtualServiceData)
}

func (a *appmeshTranslationDao) EnsureVirtualRouter(
	mesh *smh_discovery.Mesh,
	virtualRouter *appmesh.VirtualRouterData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	return appmeshClient.EnsureVirtualRouter(virtualRouter)
}

func (a *appmeshTranslationDao) EnsureRoute(
	mesh *smh_discovery.Mesh,
	route *appmesh.RouteData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	return appmeshClient.EnsureRoute(route)
}

func (a *appmeshTranslationDao) EnsureVirtualNode(
	mesh *smh_discovery.Mesh,
	virtualNode *appmesh.VirtualNodeData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	return appmeshClient.EnsureVirtualNode(virtualNode)
}

func (a *appmeshTranslationDao) ReconcileVirtualRoutersAndRoutesAndVirtualServices(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
	virtualRouters []*appmesh.VirtualRouterData,
	routes []*appmesh.RouteData,
	virtualServices []*appmesh.VirtualServiceData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	meshName := aws2.String(mesh.Spec.GetAwsAppMesh().GetName())
	return appmeshClient.ReconcileVirtualRoutersAndRoutesAndVirtualServices(ctx, meshName, virtualRouters, routes, virtualServices)
}

func (a *appmeshTranslationDao) ReconcileVirtualNodes(
	ctx context.Context,
	mesh *smh_discovery.Mesh,
	virtualNodes []*appmesh.VirtualNodeData,
) error {
	appmeshClient, err := a.appmeshClientFactory(mesh)
	if err != nil {
		return err
	}
	meshName := aws2.String(mesh.Spec.GetAwsAppMesh().GetName())
	return appmeshClient.ReconcileVirtualNodes(ctx, meshName, virtualNodes)
}

func isServiceBackedByWorkload(
	meshService *smh_discovery.MeshService,
	meshWorkload *smh_discovery.MeshWorkload,
) bool {
	if meshService.Spec.GetKubeService().GetRef().GetNamespace() != meshWorkload.Spec.GetKubeController().GetKubeControllerRef().GetNamespace() ||
		len(meshService.Spec.GetKubeService().GetWorkloadSelectorLabels()) == 0 ||
		len(meshWorkload.Spec.GetKubeController().GetLabels()) == 0 {
		return false
	}
	return labels.AreLabelsInWhiteList(
		meshService.Spec.GetKubeService().GetWorkloadSelectorLabels(),
		meshWorkload.Spec.GetKubeController().GetLabels(),
	)
}
