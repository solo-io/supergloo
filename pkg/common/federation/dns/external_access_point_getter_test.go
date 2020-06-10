package dns_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	. "github.com/solo-io/go-utils/testutils"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/federation/dns"
	istio_federation "github.com/solo-io/service-mesh-hub/pkg/common/federation/resolver/meshes/istio"
	mock_multicluster "github.com/solo-io/service-mesh-hub/pkg/common/kube/multicluster/mocks"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("external access point getter", func() {
	var (
		ctrl                      *gomock.Controller
		ctx                       context.Context
		nodeClient                *mock_kubernetes_core.MockNodeClient
		podClient                 *mock_kubernetes_core.MockPodClient
		dynamicClientGetter       *mock_multicluster.MockDynamicClientGetter
		externalAccessPointGetter dns.ExternalAccessPointGetter

		clusterName = "test-cluster"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		podClient = mock_kubernetes_core.NewMockPodClient(ctrl)
		nodeClient = mock_kubernetes_core.NewMockNodeClient(ctrl)
		dynamicClientGetter = mock_multicluster.NewMockDynamicClientGetter(ctrl)
		externalAccessPointGetter = dns.NewExternalAccessPointGetter(
			dynamicClientGetter,
			func(client client.Client) k8s_core.PodClient {
				return podClient
			},
			func(client client.Client) k8s_core.NodeClient {
				return nodeClient
			},
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("will return an error if service has no ports", func() {
		svc := &corev1.Service{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
		}
		_, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(dns.NoAvailablePorts(svc, clusterName)))
	})

	It("will return an error if no port can be found with the given name", func() {
		svc := &corev1.Service{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name: "incorrect-name",
					},
				},
			},
		}
		_, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(dns.NamedPortNotFound(svc, clusterName, istio_federation.DefaultGatewayPortName)))
	})

	It("will return an error if service does not match the given types", func() {

		svc := &corev1.Service{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name: istio_federation.DefaultGatewayPortName,
					},
				},
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		_, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
		Expect(err).To(HaveOccurred())
		Expect(err).To(HaveInErrorChain(dns.UnsupportedServiceType(svc, clusterName)))
	})

	Context("node port", func() {

		It("will return an error if dynamicClient cannot be found", func() {

			svc := &corev1.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: istio_federation.DefaultGatewayPortName,
						},
					},
					Type: corev1.ServiceTypeNodePort,
				},
			}
			noClientForClusterError := eris.New("")
			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName).
				Return(nil, noClientForClusterError)

			_, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(noClientForClusterError))
		})

		It("will return an error if no pods are scheduled", func() {

			svc := &corev1.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: istio_federation.DefaultGatewayPortName,
						},
					},
					Type: corev1.ServiceTypeNodePort,
				},
			}
			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName).
				Return(nil, nil)

			podClient.EXPECT().
				ListPod(ctx, &client.ListOptions{
					LabelSelector: labels.SelectorFromSet(svc.Spec.Selector),
					Namespace:     svc.Namespace,
				}).
				Return(&corev1.PodList{}, nil)

			_, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(dns.NoScheduledPods(svc, clusterName)))
		})

		It("will return an error if the node has no valid addresses", func() {

			svc := &corev1.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: istio_federation.DefaultGatewayPortName,
						},
					},
					Type: corev1.ServiceTypeNodePort,
				},
			}
			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName).
				Return(nil, nil)

			nodeName := "test-node"
			podClient.EXPECT().
				ListPod(ctx, &client.ListOptions{
					LabelSelector: labels.SelectorFromSet(svc.Spec.Selector),
					Namespace:     svc.Namespace,
				}).
				Return(&corev1.PodList{Items: []corev1.Pod{
					{
						Spec: corev1.PodSpec{
							NodeName: nodeName,
						},
					},
				}}, nil)

			node := &corev1.Node{}
			nodeClient.EXPECT().
				GetNode(ctx, client.ObjectKey{Name: nodeName}).
				Return(node, nil)

			_, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(dns.NoActiveAddressesForNode(node, clusterName)))
		})

		It("will return address from node, and port from svc", func() {

			svc := &corev1.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:     istio_federation.DefaultGatewayPortName,
							NodePort: 32000,
						},
					},
					Type: corev1.ServiceTypeNodePort,
				},
			}
			dynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, clusterName).
				Return(nil, nil)

			nodeName := "test-node"
			podClient.EXPECT().
				ListPod(ctx, &client.ListOptions{
					LabelSelector: labels.SelectorFromSet(svc.Spec.Selector),
					Namespace:     svc.Namespace,
				}).
				Return(&corev1.PodList{Items: []corev1.Pod{
					{
						Spec: corev1.PodSpec{
							NodeName: nodeName,
						},
					},
				}}, nil)

			node := &corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{
							Address: "fake-address",
						},
					},
				},
			}
			nodeClient.EXPECT().
				GetNode(ctx, client.ObjectKey{Name: nodeName}).
				Return(node, nil)

			eap, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(eap.Port).To(Equal(uint32(svc.Spec.Ports[0].NodePort)))
			Expect(eap.Address).To(Equal(node.Status.Addresses[0].Address))
		})

	})

	Context("load balancer", func() {

		It("will return an error if no load balancers are available", func() {

			svc := &corev1.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: istio_federation.DefaultGatewayPortName,
						},
					},
					Type: corev1.ServiceTypeLoadBalancer,
				},
			}
			_, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(dns.NoAvailableIngresses(svc, clusterName)))
		})

		It("will return an error if no externally resolvable IP is available", func() {

			svc := &corev1.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: istio_federation.DefaultGatewayPortName,
						},
					},
					Type: corev1.ServiceTypeLoadBalancer,
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{}},
					},
				},
			}
			_, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(dns.NoExternallyResolvableIp(svc, clusterName)))
		})

		It("will return an error if no fqdn is available on the ingress", func() {

			svc := &corev1.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: istio_federation.DefaultGatewayPortName,
						},
					},
					Type: corev1.ServiceTypeLoadBalancer,
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{}},
					},
				},
			}
			_, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
			Expect(err).To(HaveOccurred())
			Expect(err).To(HaveInErrorChain(dns.NoExternallyResolvableIp(svc, clusterName)))
		})

		It("will return ingress ip and port", func() {
			ip := "0.0.0.0"
			svc := &corev1.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: istio_federation.DefaultGatewayPortName,
							Port: 32000,
						},
					},
					Type: corev1.ServiceTypeLoadBalancer,
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{
							IP: ip,
						}},
					},
				},
			}
			eap, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(eap.Port).To(Equal(uint32(svc.Spec.Ports[0].Port)))
			Expect(eap.Address).To(Equal(ip))
		})

		It("will return ingress hostname and port", func() {
			hostName := "test.host.com"
			svc := &corev1.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: istio_federation.DefaultGatewayPortName,
							Port: 32000,
						},
					},
					Type: corev1.ServiceTypeLoadBalancer,
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{
							Hostname: hostName,
						}},
					},
				},
			}
			eap, err := externalAccessPointGetter.GetExternalAccessPointForService(ctx, svc, istio_federation.DefaultGatewayPortName, clusterName)
			Expect(err).NotTo(HaveOccurred())
			Expect(eap.Port).To(Equal(uint32(svc.Spec.Ports[0].Port)))
			Expect(eap.Address).To(Equal(hostName))
		})

	})
})
