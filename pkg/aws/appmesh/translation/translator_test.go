package translation_test

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/aws/appmesh/translation"
	"github.com/solo-io/service-mesh-hub/pkg/metadata"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Translator", func() {
	var (
		ctrl              *gomock.Controller
		appmeshTranslator translation.AppmeshTranslator
		appmeshName       = aws.String("appmesh-name")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		appmeshTranslator = translation.NewAppmeshTranslator()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should build VirtualNode", func() {
		meshWorkload := &v1alpha1.MeshWorkload{
			ObjectMeta: v1.ObjectMeta{Name: "workload-name", Namespace: "workload-namespace"},
			Spec: types.MeshWorkloadSpec{
				Appmesh: &types.MeshWorkloadSpec_Appmesh{
					VirtualNodeName: "virtual-node-name",
					Ports: []*types.MeshWorkloadSpec_Appmesh_ContainerPort{
						{
							Port:     9080,
							Protocol: "tcp",
						},
						{
							Port:     9081,
							Protocol: "udp",
						},
					},
				},
			},
		}
		meshService := &v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{Name: "service-name", Namespace: "service-namespace"},
			Spec: types.MeshServiceSpec{
				KubeService: &types.MeshServiceSpec_KubeService{
					Ref: &types2.ResourceRef{
						Name: "service-name",
					},
				},
			},
		}
		upstreamServices := []*v1alpha1.MeshService{
			{
				Spec: types.MeshServiceSpec{
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: &types2.ResourceRef{
							Name: "upstream-service-nameA",
						},
					},
				},
			},
			{
				Spec: types.MeshServiceSpec{
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: &types2.ResourceRef{
							Name: "upstream-service-nameB",
						},
					},
				},
			},
		}
		listeners := []*appmesh.Listener{
			{
				PortMapping: &appmesh.PortMapping{
					Port: aws.Int64(9080),
					//Protocol: aws.String("tcp"),
					Protocol: aws.String("http"),
				},
			},
			{
				PortMapping: &appmesh.PortMapping{
					Port: aws.Int64(9081),
					//Protocol: aws.String("udp"),
					Protocol: aws.String("http"),
				},
			},
		}
		backends := []*appmesh.Backend{
			{
				VirtualService: &appmesh.VirtualServiceBackend{
					VirtualServiceName: aws.String(metadata.BuildVirtualServiceName(upstreamServices[0])),
				},
			},
			{
				VirtualService: &appmesh.VirtualServiceBackend{
					VirtualServiceName: aws.String(metadata.BuildVirtualServiceName(upstreamServices[1])),
				},
			},
		}
		expectedVirtualNode := &appmesh.VirtualNodeData{
			VirtualNodeName: aws.String(metadata.BuildVirtualNodeName(meshWorkload)),
			MeshName:        appmeshName,
			Spec: &appmesh.VirtualNodeSpec{
				Listeners: listeners,
				Backends:  backends,
				ServiceDiscovery: &appmesh.ServiceDiscovery{
					Dns: &appmesh.DnsServiceDiscovery{
						Hostname: aws.String(metadata.BuildLocalFQDN(meshService.Spec.GetKubeService().GetRef().GetName())),
					},
				},
			},
		}
		virtualNode := appmeshTranslator.BuildVirtualNode(appmeshName, meshWorkload, meshService, upstreamServices)
		Expect(virtualNode).To(Equal(expectedVirtualNode))
	})

	It("should build Route", func() {
		routeName := "route-name"
		priority := 0
		meshService := &v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{Name: "service-name", Namespace: "service-namespace"},
			Spec: types.MeshServiceSpec{
				KubeService: &types.MeshServiceSpec_KubeService{
					Ref: &types2.ResourceRef{
						Name: "upstream-service-nameA",
					},
				},
			},
		}
		meshWorkloads := []*v1alpha1.MeshWorkload{
			{
				Spec: types.MeshWorkloadSpec{
					Appmesh: &types.MeshWorkloadSpec_Appmesh{
						VirtualNodeName: "virtual-node-name-A",
					},
				},
			},
			{
				Spec: types.MeshWorkloadSpec{
					Appmesh: &types.MeshWorkloadSpec_Appmesh{
						VirtualNodeName: "virtual-node-name-B",
					},
				},
			},
		}
		weightedTargets := []*appmesh.WeightedTarget{
			{
				VirtualNode: aws.String(metadata.BuildVirtualNodeName(meshWorkloads[0])),
				Weight:      aws.Int64(1),
			},
			{
				VirtualNode: aws.String(metadata.BuildVirtualNodeName(meshWorkloads[1])),
				Weight:      aws.Int64(1),
			},
		}
		expectedRoute := &appmesh.RouteData{
			MeshName:  appmeshName,
			RouteName: aws.String(routeName),
			Spec: &appmesh.RouteSpec{
				HttpRoute: &appmesh.HttpRoute{
					Action: &appmesh.HttpRouteAction{
						WeightedTargets: weightedTargets,
					},
					Match: &appmesh.HttpRouteMatch{
						Prefix: aws.String("/"),
					},
				},
				Priority: aws.Int64(int64(priority)),
			},
			VirtualRouterName: aws.String(metadata.BuildVirtualRouterName(meshService)),
		}
		route, err := appmeshTranslator.BuildRoute(appmeshName, routeName, priority, meshService, meshWorkloads)
		Expect(err).ToNot(HaveOccurred())
		Expect(route).To(Equal(expectedRoute))
	})

	It("should return error if targets more than 10 workloads when building Route", func() {
		routeName := "route-name"
		priority := 0
		meshService := &v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
		}
		meshWorkloads := []*v1alpha1.MeshWorkload{
			{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {},
		}
		route, err := appmeshTranslator.BuildRoute(appmeshName, routeName, priority, meshService, meshWorkloads)
		Expect(err).To(testutils.HaveInErrorChain(translation.ExceededMaximumWorkloadsError(meshService)))
		Expect(route).To(BeNil())
	})

	It("should build VirtualService", func() {
		meshService := &v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{Name: "service-name", Namespace: "service-namespace"},
			Spec: types.MeshServiceSpec{
				KubeService: &types.MeshServiceSpec_KubeService{
					Ref: &types2.ResourceRef{
						Name: "upstream-service-nameA",
					},
				},
			},
		}
		expectedVirtualService := &appmesh.VirtualServiceData{
			MeshName: appmeshName,
			Spec: &appmesh.VirtualServiceSpec{
				Provider: &appmesh.VirtualServiceProvider{
					VirtualRouter: &appmesh.VirtualRouterServiceProvider{
						VirtualRouterName: aws.String(metadata.BuildVirtualRouterName(meshService)),
					},
				},
			},
			VirtualServiceName: aws.String(metadata.BuildVirtualServiceName(meshService)),
		}
		virtualService := appmeshTranslator.BuildVirtualService(appmeshName, meshService)
		Expect(virtualService).To(Equal(expectedVirtualService))
	})

	It("should build VirtualRouter", func() {
		meshService := &v1alpha1.MeshService{
			ObjectMeta: v1.ObjectMeta{Name: "service-name", Namespace: "service-namespace"},
			Spec: types.MeshServiceSpec{
				KubeService: &types.MeshServiceSpec_KubeService{
					Ref: &types2.ResourceRef{
						Name: "upstream-service-nameA",
					},
					Ports: []*types.MeshServiceSpec_KubeService_KubeServicePort{
						{
							Port:     9080,
							Protocol: "tcp",
						},
						{
							Port:     9081,
							Protocol: "udp",
						},
					},
				},
			},
		}
		virtualRouterListeners := []*appmesh.VirtualRouterListener{
			{
				PortMapping: &appmesh.PortMapping{
					Port: aws.Int64(int64(meshService.Spec.GetKubeService().GetPorts()[0].GetPort())),
					//Protocol: aws.String(meshService.Spec.GetKubeService().GetPorts()[0].GetProtocol()),
					Protocol: aws.String("http"),
				},
			},
			{
				PortMapping: &appmesh.PortMapping{
					Port: aws.Int64(int64(meshService.Spec.GetKubeService().GetPorts()[1].GetPort())),
					//Protocol: aws.String(meshService.Spec.GetKubeService().GetPorts()[1].GetProtocol()),
					Protocol: aws.String("http"),
				},
			},
		}
		expectedVirtualRouter := &appmesh.VirtualRouterData{
			MeshName:          appmeshName,
			VirtualRouterName: aws.String(metadata.BuildVirtualRouterName(meshService)),
			Spec: &appmesh.VirtualRouterSpec{
				Listeners: virtualRouterListeners,
			},
		}
		virtualRouter := appmeshTranslator.BuildVirtualRouter(appmeshName, meshService)
		Expect(virtualRouter).To(Equal(expectedVirtualRouter))
	})
})
