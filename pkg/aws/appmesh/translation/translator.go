package translation

import (
	aws2 "github.com/aws/aws-sdk-go/aws"
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/rotisserie/eris"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
)

var (
	ExceededMaximumWorkloadsError = func(meshService *zephyr_discovery.MeshService) error {
		return eris.Errorf("Workloads selected by service %s.%s exceeds Appmesh's maximum of 10 weighted targets.",
			meshService.GetName(), meshService.GetNamespace())
	}
)

type appmeshTranslator struct {
}

func NewAppmeshTranslator() AppmeshTranslator {
	return &appmeshTranslator{}
}

func (a *appmeshTranslator) BuildVirtualNode(
	appmeshName *string,
	meshWorkload *zephyr_discovery.MeshWorkload,
	meshService *zephyr_discovery.MeshService,
	upstreamServices []*zephyr_discovery.MeshService,
) *appmesh2.VirtualNodeData {
	virtualNodeName := aws2.String(metadata.BuildVirtualNodeName(meshWorkload))
	var backends []*appmesh2.Backend
	for _, upstreamService := range upstreamServices {
		backends = append(backends, &appmesh2.Backend{
			VirtualService: &appmesh2.VirtualServiceBackend{
				VirtualServiceName: aws2.String(metadata.BuildVirtualServiceName(upstreamService)),
			},
		})
	}
	var listeners []*appmesh2.Listener
	for _, containerPort := range meshWorkload.Spec.GetAppmesh().GetPorts() {
		listeners = append(listeners, &appmesh2.Listener{
			PortMapping: &appmesh2.PortMapping{
				Port:     aws2.Int64(int64(containerPort.Port)),
				Protocol: aws2.String(containerPort.Protocol),
			},
		})
	}
	return &appmesh2.VirtualNodeData{
		VirtualNodeName: virtualNodeName,
		MeshName:        appmeshName,
		Spec: &appmesh2.VirtualNodeSpec{
			Listeners: listeners,
			Backends:  backends,
			ServiceDiscovery: &appmesh2.ServiceDiscovery{
				Dns: &appmesh2.DnsServiceDiscovery{
					Hostname: aws2.String(metadata.BuildLocalFQDN(meshService.GetName())),
				},
			},
		},
	}
}

func (a *appmeshTranslator) BuildRoute(
	appmeshName *string,
	routeName string,
	priority int,
	meshService *zephyr_discovery.MeshService,
	meshWorkloads []*zephyr_discovery.MeshWorkload,
) (*appmesh2.RouteData, error) {
	if len(meshWorkloads) > 10 {
		return nil, ExceededMaximumWorkloadsError(meshService)
	}
	var weightedTargets []*appmesh2.WeightedTarget
	for _, meshWorkload := range meshWorkloads {
		weightedTargets = append(weightedTargets, &appmesh2.WeightedTarget{
			VirtualNode: aws2.String(metadata.BuildVirtualNodeName(meshWorkload)),
			Weight:      aws2.Int64(1),
		})
	}
	return &appmesh2.RouteData{
		MeshName:  appmeshName,
		RouteName: aws2.String(routeName),
		Spec: &appmesh2.RouteSpec{
			HttpRoute: &appmesh2.HttpRoute{
				Action: &appmesh2.HttpRouteAction{
					WeightedTargets: weightedTargets,
				},
				Match: &appmesh2.HttpRouteMatch{
					Prefix: aws2.String("/"),
				},
			},
			Priority: aws2.Int64(int64(priority)),
		},
		VirtualRouterName: aws2.String(metadata.BuildVirtualRouterName(meshService)),
	}, nil
}

func (a *appmeshTranslator) BuildVirtualService(
	appmeshName *string,
	meshService *zephyr_discovery.MeshService,
) *appmesh2.VirtualServiceData {
	return &appmesh2.VirtualServiceData{
		MeshName: appmeshName,
		Spec: &appmesh2.VirtualServiceSpec{
			Provider: &appmesh2.VirtualServiceProvider{
				VirtualRouter: &appmesh2.VirtualRouterServiceProvider{
					VirtualRouterName: aws2.String(metadata.BuildVirtualRouterName(meshService)),
				},
			},
		},
		VirtualServiceName: aws2.String(metadata.BuildVirtualServiceName(meshService)),
	}
}

func (a *appmeshTranslator) BuildVirtualRouter(
	appmeshName *string,
	meshService *zephyr_discovery.MeshService,
) *appmesh2.VirtualRouterData {
	var virtualRouterListeners []*appmesh2.VirtualRouterListener
	for _, servicePort := range meshService.Spec.GetKubeService().GetPorts() {
		virtualRouterListeners = append(virtualRouterListeners, &appmesh2.VirtualRouterListener{
			PortMapping: &appmesh2.PortMapping{
				Port:     aws2.Int64(int64(servicePort.GetPort())),
				Protocol: aws2.String(servicePort.GetProtocol()),
			},
		})
	}
	return &appmesh2.VirtualRouterData{
		MeshName:          appmeshName,
		VirtualRouterName: aws2.String(metadata.BuildVirtualRouterName(meshService)),
		Spec: &appmesh2.VirtualRouterSpec{
			Listeners: virtualRouterListeners,
		},
	}
}
