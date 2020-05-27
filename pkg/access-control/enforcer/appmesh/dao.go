package appmesh

import (
	"context"
	"strings"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	"github.com/rotisserie/eris"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/aws"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/selector"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/aws/clients/appmesh"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	// Canonical name for default route that permits traffic to all workloads backing service with equal weight.
	DefaultRouteName = "default"
)

var (
	ExceededMaximumWorkloadsError = func(meshService *zephyr_discovery.MeshService) error {
		return eris.Errorf("Workloads selected by service %s.%s exceeds Appmesh's maximum of 10 weighted targets.",
			meshService.GetName(), meshService.GetNamespace())
	}
)

type appmeshAccessControlDao struct {
	meshServiceClient    zephyr_discovery.MeshServiceClient
	meshWorkloadClient   zephyr_discovery.MeshWorkloadClient
	resourceSelector     selector.ResourceSelector
	appmeshClientFactory appmesh.AppMeshClientFactory
	awsCredentialsGetter aws.AwsCredentialsGetter
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

func (a *appmeshAccessControlDao) EnsureVirtualServicesWithDefaultRoutes(
	mesh *zephyr_discovery.Mesh,
	serviceToWorkloads map[*zephyr_discovery.MeshService][]*zephyr_discovery.MeshWorkload,
) error {
	appmeshClient, err := a.buildAppmeshClient(mesh)
	if err != nil {
		return err
	}
	for service, workloads := range serviceToWorkloads {
		meshName := aws2.String(mesh.Spec.GetAwsAppMesh().GetName())
		virtualServiceName := aws2.String(metadata.BuildVirtualServiceName(service))
		virtualRouterName := aws2.String(metadata.BuildVirtualRouterName(service))
		virtualNodeNames, err := a.fetchVirtualNodeNamesForWorkloads(workloads)
		if err != nil {
			return err
		}
		if len(virtualNodeNames) > 10 {
			return ExceededMaximumWorkloadsError(service)
		}
		err = a.ensureVirtualService(appmeshClient, meshName, virtualServiceName, virtualRouterName)
		if err != nil {
			return err
		}
		err = a.ensureVirtualRouter(appmeshClient, meshName, virtualRouterName, service.Spec.GetKubeService().GetPorts())
		if err != nil {
			return err
		}
		err = a.ensureRouteToAllWorkloads(appmeshClient, meshName, virtualRouterName, virtualNodeNames)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *appmeshAccessControlDao) EnsureVirtualNodesWithDefaultBackends(
	mesh *zephyr_discovery.Mesh,
	workloadToServices map[*zephyr_discovery.MeshWorkload][]*zephyr_discovery.MeshService,
) error {
	appmeshClient, err := a.buildAppmeshClient(mesh)
	if err != nil {
		return err
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

func (a *appmeshAccessControlDao) ensureVirtualService(
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

func (a *appmeshAccessControlDao) fetchVirtualNodeNamesForWorkloads(
	meshWorkloads []*zephyr_discovery.MeshWorkload,
) ([]string, error) {
	var virtualNodeNames []string
	for _, meshWorkload := range meshWorkloads {
		virtualNodeNames = append(virtualNodeNames, metadata.BuildVirtualNodeName(meshWorkload))
	}
	return virtualNodeNames, nil
}

// For a VirtualRouter, ensure a canonical default route that:
//      1. routes to all workload backing the given service with equal weight
//      2. has priority 0
func (a *appmeshAccessControlDao) ensureRouteToAllWorkloads(
	appmeshClient appmeshiface.AppMeshAPI,
	meshName *string,
	virtualRouterName *string,
	virtualNodeNames []string,
) error {
	req := &appmesh2.ListRoutesInput{
		MeshName:          meshName,
		VirtualRouterName: virtualRouterName,
	}
	var routeNames []*string
	for {
		resp, err := appmeshClient.ListRoutes(req)
		if err != nil {
			return err
		}
		for _, routeRef := range resp.Routes {
			routeNames = append(routeNames, routeRef.RouteName)
		}
		if req.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}
	// Handle case where there are multiple default routes defined.
	var extantDefaultRoutes []*appmesh2.RouteData
	for _, routeName := range routeNames {
		if aws2.StringValue(routeName) != DefaultRouteName {
			continue
		}
		resp, err := appmeshClient.DescribeRoute(&appmesh2.DescribeRouteInput{
			MeshName:          meshName,
			RouteName:         routeName,
			VirtualRouterName: virtualRouterName,
		})
		if err != nil {
			return err
		}
		extantDefaultRoutes = append(extantDefaultRoutes, resp.Route)
	}
	canonicalDefaultRoute := buildDefaultRouteInput(meshName, virtualRouterName, virtualNodeNames)
	// Maintain a route if it's equivalent to the canonical default route, otherwise delete.
	// If no extant routes match the default, create one.
	if len(extantDefaultRoutes) > 1 {
		var routesToDelete []*appmesh2.DeleteRouteInput
		foundMatchingDefaultRoute := false
		for _, extantDefaultRoute := range extantDefaultRoutes {
			if !foundMatchingDefaultRoute && areRoutesEqual(extantDefaultRoute, canonicalDefaultRoute) {
				foundMatchingDefaultRoute = true
			} else {
				routesToDelete = append(routesToDelete, &appmesh2.DeleteRouteInput{
					MeshName:          meshName,
					RouteName:         extantDefaultRoute.VirtualRouterName,
					VirtualRouterName: virtualRouterName,
				})
			}
		}
		for _, routeToDelete := range routesToDelete {
			_, err := appmeshClient.DeleteRoute(routeToDelete)
			if err != nil {
				return err
			}
		}
		if foundMatchingDefaultRoute {
			return nil
		}
	} else if len(extantDefaultRoutes) == 1 {
		if areRoutesEqual(extantDefaultRoutes[0], canonicalDefaultRoute) {
			return nil
		}
		_, err := appmeshClient.DeleteRoute(&appmesh2.DeleteRouteInput{
			MeshName:          meshName,
			RouteName:         extantDefaultRoutes[0].VirtualRouterName,
			VirtualRouterName: virtualRouterName,
		})
		if err != nil {
			return err
		}
	}
	_, err := appmeshClient.CreateRoute(canonicalDefaultRoute)
	if err != nil {
		return err
	}
	return nil
}

func buildDefaultRouteInput(
	meshName *string,
	virtualRouterName *string,
	virtualNodeNames []string,
) *appmesh2.CreateRouteInput {
	var weightedTargets []*appmesh2.WeightedTarget
	for _, virtualNodeName := range virtualNodeNames {
		weightedTargets = append(weightedTargets, &appmesh2.WeightedTarget{
			VirtualNode: aws2.String(virtualNodeName),
			Weight:      aws2.Int64(1),
		})
	}
	return &appmesh2.CreateRouteInput{
		MeshName:  meshName,
		RouteName: aws2.String(DefaultRouteName),
		Spec: &appmesh2.RouteSpec{
			HttpRoute: &appmesh2.HttpRoute{
				Action: &appmesh2.HttpRouteAction{
					WeightedTargets: weightedTargets,
				},
				Match: &appmesh2.HttpRouteMatch{
					Prefix: aws2.String("/"),
				},
			},
			Priority: aws2.Int64(0),
		},
		VirtualRouterName: virtualRouterName,
	}
}

func areRoutesEqual(routeA *appmesh2.RouteData, routeB *appmesh2.CreateRouteInput) bool {
	if aws2.StringValue(routeA.RouteName) != aws2.StringValue(routeB.RouteName) ||
		aws2.StringValue(routeA.MeshName) != aws2.StringValue(routeB.MeshName) ||
		aws2.StringValue(routeA.VirtualRouterName) != aws2.StringValue(routeB.VirtualRouterName) ||
		aws2.Int64Value(routeA.Spec.Priority) != aws2.Int64Value(routeB.Spec.Priority) ||
		routeA.Spec.GrpcRoute != routeB.Spec.GrpcRoute ||
		routeA.Spec.Http2Route != routeB.Spec.Http2Route ||
		routeA.Spec.TcpRoute != routeB.Spec.TcpRoute ||
		routeA.Spec.HttpRoute.RetryPolicy != routeB.Spec.HttpRoute.RetryPolicy ||
		routeA.Spec.HttpRoute.Match.Scheme != routeB.Spec.HttpRoute.Match.Scheme ||
		routeA.Spec.HttpRoute.Match.Prefix != routeB.Spec.HttpRoute.Match.Prefix ||
		routeA.Spec.HttpRoute.Match.Method != routeB.Spec.HttpRoute.Match.Method ||
		routeA.Spec.HttpRoute.Match.Headers != nil ||
		routeB.Spec.HttpRoute.Match.Headers != nil {
		return false
	}
	weightedTargetsA := map[string]int64{}
	for _, weightedTargetA := range routeA.Spec.HttpRoute.Action.WeightedTargets {
		weightedTargetsA[aws2.StringValue(weightedTargetA.VirtualNode)] = aws2.Int64Value(weightedTargetA.Weight)
	}
	for _, weightedTargetB := range routeB.Spec.HttpRoute.Action.WeightedTargets {
		virtualNodeNameB := aws2.StringValue(weightedTargetB.VirtualNode)
		weightA, ok := weightedTargetsA[virtualNodeNameB]
		if !ok || weightA != aws2.Int64Value(weightedTargetB.Weight) {
			return false
		}
		delete(weightedTargetsA, virtualNodeNameB)
	}
	return len(weightedTargetsA) == 0
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
