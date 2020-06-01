package appmesh_test

import (
	aws2 "github.com/aws/aws-sdk-go/aws"
	appmesh2 "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/service-mesh-hub/pkg/aws/appmesh"
)

var _ = Describe("Matcher", func() {
	var (
		ctrl           *gomock.Controller
		meshName       = aws2.String("mesh-name")
		appmeshMatcher AppmeshMatcher
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		appmeshMatcher = NewAppmeshMatcher()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should return true if Routes are equal", func() {
		vrName := aws2.String("virtual-router-name")
		routeA := &appmesh2.RouteData{
			MeshName:          meshName,
			VirtualRouterName: vrName,
			RouteName:         aws2.String("routeA"),
			Spec: &appmesh2.RouteSpec{
				HttpRoute: &appmesh2.HttpRoute{
					Action: &appmesh2.HttpRouteAction{
						WeightedTargets: []*appmesh2.WeightedTarget{
							{
								VirtualNode: aws2.String("vn-1"),
								Weight:      aws2.Int64(1),
							},
							{
								VirtualNode: aws2.String("vn-2"),
								Weight:      aws2.Int64(2),
							},
							{
								VirtualNode: aws2.String("vn-3"),
								Weight:      aws2.Int64(3),
							},
						},
					},
					Match: &appmesh2.HttpRouteMatch{
						Method: aws2.String("GET"),
						Prefix: aws2.String("/"),
					},
				},
				Priority: aws2.Int64(0),
			},
		}
		routeB := &appmesh2.RouteData{
			MeshName:          meshName,
			VirtualRouterName: vrName,
			RouteName:         aws2.String("routeA"),
			Spec: &appmesh2.RouteSpec{
				HttpRoute: &appmesh2.HttpRoute{
					Action: &appmesh2.HttpRouteAction{
						WeightedTargets: []*appmesh2.WeightedTarget{
							{
								VirtualNode: aws2.String("vn-2"),
								Weight:      aws2.Int64(2),
							},
							{
								VirtualNode: aws2.String("vn-1"),
								Weight:      aws2.Int64(1),
							},
							{
								VirtualNode: aws2.String("vn-3"),
								Weight:      aws2.Int64(3),
							},
						},
					},
					Match: &appmesh2.HttpRouteMatch{
						Method: aws2.String("GET"),
						Prefix: aws2.String("/"),
					},
				},
				Priority: aws2.Int64(0),
			},
		}
		equal := appmeshMatcher.AreRoutesEqual(routeA, routeB)
		Expect(equal).To(BeTrue())
	})

	It("should return false if Routes are not equal", func() {
		vrName := aws2.String("virtual-router-name")
		routeA := &appmesh2.RouteData{
			MeshName:          meshName,
			VirtualRouterName: vrName,
			RouteName:         aws2.String("routeA"),
			Spec: &appmesh2.RouteSpec{
				HttpRoute: &appmesh2.HttpRoute{
					Action: &appmesh2.HttpRouteAction{
						WeightedTargets: []*appmesh2.WeightedTarget{
							{
								VirtualNode: aws2.String("vn-1"),
								Weight:      aws2.Int64(1),
							},
							{
								VirtualNode: aws2.String("vn-2"),
								Weight:      aws2.Int64(2),
							},
						},
					},
					Match: &appmesh2.HttpRouteMatch{
						Method: aws2.String("GET"),
						Prefix: aws2.String("/"),
					},
				},
				Priority: aws2.Int64(0),
			},
		}
		routeB := &appmesh2.RouteData{
			MeshName:          meshName,
			VirtualRouterName: vrName,
			RouteName:         aws2.String("routeA"),
			Spec: &appmesh2.RouteSpec{
				HttpRoute: &appmesh2.HttpRoute{
					Action: &appmesh2.HttpRouteAction{
						WeightedTargets: []*appmesh2.WeightedTarget{
							{
								VirtualNode: aws2.String("vn-2"),
								Weight:      aws2.Int64(2),
							},
							{
								VirtualNode: aws2.String("vn-1"),
								Weight:      aws2.Int64(1),
							},
							{
								VirtualNode: aws2.String("vn-3"),
								Weight:      aws2.Int64(3),
							},
						},
					},
					Match: &appmesh2.HttpRouteMatch{
						Method: aws2.String("GET"),
						Prefix: aws2.String("/"),
					},
				},
				Priority: aws2.Int64(0),
			},
		}
		equal := appmeshMatcher.AreRoutesEqual(routeA, routeB)
		Expect(equal).To(BeFalse())
	})

	It("should return true if VirtualNodes are equal", func() {
		vnA := &appmesh2.VirtualNodeData{
			MeshName:        meshName,
			VirtualNodeName: aws2.String("vn-name"),
			Spec: &appmesh2.VirtualNodeSpec{
				Backends: []*appmesh2.Backend{
					{
						VirtualService: &appmesh2.VirtualServiceBackend{
							VirtualServiceName: aws2.String("vs-name1"),
						},
					},
					{
						VirtualService: &appmesh2.VirtualServiceBackend{
							VirtualServiceName: aws2.String("vs-name2"),
						},
					},
				},
				Listeners: []*appmesh2.Listener{
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9080),
							Protocol: aws2.String("TCP"),
						},
					},
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9081),
							Protocol: aws2.String("UDP"),
						},
					},
				},
				ServiceDiscovery: &appmesh2.ServiceDiscovery{
					Dns: &appmesh2.DnsServiceDiscovery{
						Hostname: aws2.String("host-name"),
					},
				},
			},
		}
		vnB := &appmesh2.VirtualNodeData{
			MeshName:        meshName,
			VirtualNodeName: aws2.String("vn-name"),
			Spec: &appmesh2.VirtualNodeSpec{
				Backends: []*appmesh2.Backend{
					{
						VirtualService: &appmesh2.VirtualServiceBackend{
							VirtualServiceName: aws2.String("vs-name2"),
						},
					},
					{
						VirtualService: &appmesh2.VirtualServiceBackend{
							VirtualServiceName: aws2.String("vs-name1"),
						},
					},
				},
				Listeners: []*appmesh2.Listener{
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9081),
							Protocol: aws2.String("UDP"),
						},
					},
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9080),
							Protocol: aws2.String("TCP"),
						},
					},
				},
				ServiceDiscovery: &appmesh2.ServiceDiscovery{
					Dns: &appmesh2.DnsServiceDiscovery{
						Hostname: aws2.String("host-name"),
					},
				},
			},
		}
		equal := appmeshMatcher.AreVirtualNodesEqual(vnA, vnB)
		Expect(equal).To(BeTrue())
	})

	It("should return false if VirtualNodes are not equal", func() {
		vnA := &appmesh2.VirtualNodeData{
			MeshName:        meshName,
			VirtualNodeName: aws2.String("vn-name"),
			Spec: &appmesh2.VirtualNodeSpec{
				Backends: []*appmesh2.Backend{
					{
						VirtualService: &appmesh2.VirtualServiceBackend{
							VirtualServiceName: aws2.String("vs-name1"),
						},
					},
					{
						VirtualService: &appmesh2.VirtualServiceBackend{
							VirtualServiceName: aws2.String("vs-name2"),
						},
					},
				},
				Listeners: []*appmesh2.Listener{
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9080),
							Protocol: aws2.String("TCP"),
						},
					},
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9081),
							Protocol: aws2.String("UDP"),
						},
					},
				},
				ServiceDiscovery: &appmesh2.ServiceDiscovery{
					Dns: &appmesh2.DnsServiceDiscovery{
						Hostname: aws2.String("host-name"),
					},
				},
			},
		}
		vnB := &appmesh2.VirtualNodeData{
			MeshName:        meshName,
			VirtualNodeName: aws2.String("vn-name"),
			Spec: &appmesh2.VirtualNodeSpec{
				Backends: []*appmesh2.Backend{
					{
						VirtualService: &appmesh2.VirtualServiceBackend{
							VirtualServiceName: aws2.String("vs-name2"),
						},
					},
					{
						VirtualService: &appmesh2.VirtualServiceBackend{
							VirtualServiceName: aws2.String("vs-name1"),
						},
					},
				},
				Listeners: []*appmesh2.Listener{},
				ServiceDiscovery: &appmesh2.ServiceDiscovery{
					Dns: &appmesh2.DnsServiceDiscovery{
						Hostname: aws2.String("host-name"),
					},
				},
			},
		}
		equal := appmeshMatcher.AreVirtualNodesEqual(vnA, vnB)
		Expect(equal).To(BeFalse())
	})

	It("should return true if VirtualServices are equal", func() {
		vsA := &appmesh2.VirtualServiceData{
			MeshName:           meshName,
			VirtualServiceName: aws2.String("vs-name"),
			Spec: &appmesh2.VirtualServiceSpec{
				Provider: &appmesh2.VirtualServiceProvider{
					VirtualRouter: &appmesh2.VirtualRouterServiceProvider{
						VirtualRouterName: aws2.String("vr-name"),
					},
				},
			},
		}
		vsB := &appmesh2.VirtualServiceData{
			MeshName:           meshName,
			VirtualServiceName: aws2.String("vs-name"),
			Spec: &appmesh2.VirtualServiceSpec{
				Provider: &appmesh2.VirtualServiceProvider{
					VirtualRouter: &appmesh2.VirtualRouterServiceProvider{
						VirtualRouterName: aws2.String("vr-name"),
					},
				},
			},
		}
		equal := appmeshMatcher.AreVirtualServicesEqual(vsA, vsB)
		Expect(equal).To(BeTrue())
	})

	It("should return false if VirtualServices are not equal", func() {
		vsA := &appmesh2.VirtualServiceData{
			MeshName:           meshName,
			VirtualServiceName: aws2.String("vs-name"),
			Spec: &appmesh2.VirtualServiceSpec{
				Provider: &appmesh2.VirtualServiceProvider{
					VirtualRouter: &appmesh2.VirtualRouterServiceProvider{
						VirtualRouterName: aws2.String("vr-name"),
					},
				},
			},
		}
		vsB := &appmesh2.VirtualServiceData{
			MeshName:           meshName,
			VirtualServiceName: aws2.String("vs-name"),
			Spec: &appmesh2.VirtualServiceSpec{
				Provider: &appmesh2.VirtualServiceProvider{

				},
			},
		}
		equal := appmeshMatcher.AreVirtualServicesEqual(vsA, vsB)
		Expect(equal).To(BeFalse())
	})

	It("should return true if VirtualRouters are equal", func() {
		vrA := &appmesh2.VirtualRouterData{
			MeshName:          meshName,
			VirtualRouterName: aws2.String("vr-name"),
			Spec: &appmesh2.VirtualRouterSpec{
				Listeners: []*appmesh2.VirtualRouterListener{
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9080),
							Protocol: aws2.String("TCP"),
						},
					},
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9081),
							Protocol: aws2.String("UDP"),
						},
					},
				},
			},
		}
		vrB := &appmesh2.VirtualRouterData{
			MeshName:          meshName,
			VirtualRouterName: aws2.String("vr-name"),
			Spec: &appmesh2.VirtualRouterSpec{
				Listeners: []*appmesh2.VirtualRouterListener{
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9081),
							Protocol: aws2.String("UDP"),
						},
					},
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9080),
							Protocol: aws2.String("TCP"),
						},
					},
				},
			},
		}
		equal := appmeshMatcher.AreVirtualRoutersEqual(vrA, vrB)
		Expect(equal).To(BeTrue())
	})

	It("should return true if VirtualRouters are equal", func() {
		vrA := &appmesh2.VirtualRouterData{
			MeshName:          meshName,
			VirtualRouterName: aws2.String("vr-name"),
			Spec: &appmesh2.VirtualRouterSpec{
				Listeners: []*appmesh2.VirtualRouterListener{
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9080),
							Protocol: aws2.String("TCP"),
						},
					},
					{
						PortMapping: &appmesh2.PortMapping{
							Port:     aws2.Int64(9081),
							Protocol: aws2.String("UDP"),
						},
					},
				},
			},
		}
		vrB := &appmesh2.VirtualRouterData{
			MeshName:          meshName,
			VirtualRouterName: aws2.String("vr-name"),
			Spec: &appmesh2.VirtualRouterSpec{
				Listeners: []*appmesh2.VirtualRouterListener{},
			},
		}
		equal := appmeshMatcher.AreVirtualRoutersEqual(vrA, vrB)
		Expect(equal).To(BeFalse())
	})
})
