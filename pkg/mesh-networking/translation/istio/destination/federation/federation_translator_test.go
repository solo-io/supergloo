package federation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	mock_destinationrule "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/destinationrule/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/federation"
	mock_virtualservice "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/virtualservice/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/gloo-mesh/test/data"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/istio/pkg/config/protocol"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FederationTranslator", func() {
	var (
		ctx                           context.Context
		ctrl                          *gomock.Controller
		mockVirtualServiceTranslator  *mock_virtualservice.MockTranslator
		mockDestinationRuleTranslator *mock_destinationrule.MockTranslator
		mockReporter                  *mock_reporting.MockReporter
		federationTranslator          federation.Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockVirtualServiceTranslator = mock_virtualservice.NewMockTranslator(ctrl)
		mockDestinationRuleTranslator = mock_destinationrule.NewMockTranslator(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		federationTranslator = federation.NewTranslator(ctx, mockVirtualServiceTranslator, mockDestinationRuleTranslator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("translates federation resources for a federated Destination", func() {
		namespace := "namespace"
		clusterName := "cluster"

		destinationMesh := &discoveryv1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "config-namespace",
				Name:      "federated-mesh",
			},
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{Istio: &discoveryv1.MeshSpec_Istio{
					SmartDnsProxyingEnabled: true,
					Installation: &discoveryv1.MeshSpec_MeshInstallation{
						Namespace: namespace,
						Cluster:   clusterName,
						Version:   "1.8.1",
					},
					IngressGateways: []*discoveryv1.MeshSpec_Istio_IngressGatewayInfo{{
						ExternalAddress:  "mesh-gateway.dns.name",
						ExternalTlsPort:  8181,
						TlsContainerPort: 9191,
						WorkloadLabels:   map[string]string{"gatewaylabels": "righthere"},
					}},
				}},
			},
		}

		remoteMesh := &discoveryv1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "config-namespace",
				Name:      "client-mesh",
			},
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{Istio: &discoveryv1.MeshSpec_Istio{
					SmartDnsProxyingEnabled: true,
					Installation: &discoveryv1.MeshSpec_MeshInstallation{
						Namespace: "remote-namespace",
						Cluster:   "remote-cluster",
					},
				}},
			},
		}
		remoteMesh2 := &discoveryv1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "config-namespace",
				Name:      "client-mesh2",
			},
			Spec: discoveryv1.MeshSpec{
				Type: &discoveryv1.MeshSpec_Istio_{Istio: &discoveryv1.MeshSpec_Istio{
					SmartDnsProxyingEnabled: true,
					Installation: &discoveryv1.MeshSpec_MeshInstallation{
						Namespace: "remote-namespace2",
						Cluster:   "remote-cluster2",
					},
				}},
			},
		}

		destinationVirtualMesh := &networkingv1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "virtual-mesh",
				Namespace: namespace,
			},
		}

		destinationMeshRef := ezkube.MakeObjectRef(destinationMesh)
		remoteMeshRef := ezkube.MakeObjectRef(remoteMesh)
		remoteMeshRef2 := ezkube.MakeObjectRef(remoteMesh2)
		destinationVirtualMeshRef := ezkube.MakeObjectRef(destinationVirtualMesh)

		makeTrafficSplit := func(backingService *skv2corev1.ClusterObjectRef, subset map[string]string) *discoveryv1.DestinationStatus_AppliedTrafficPolicy {
			return &discoveryv1.DestinationStatus_AppliedTrafficPolicy{Spec: &data.RemoteTrafficShiftPolicy(
				"",
				"",
				backingService,
				clusterName,
				// NOTE(ilackarms): we only care about the subset labels here
				subset,
				0,
			).Spec}
		}

		backingService := &skv2corev1.ClusterObjectRef{
			Name:        "some-svc",
			Namespace:   "some-ns",
			ClusterName: clusterName,
		}

		destination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{KubeService: &discoveryv1.DestinationSpec_KubeService{
					Ref: backingService,
					Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
						{
							Port:     1234,
							Name:     "http",
							Protocol: "TCP",
						},
					},
				}},
				Mesh: destinationMeshRef,
			},
			// include some applied subsets
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*discoveryv1.DestinationStatus_AppliedTrafficPolicy{
					makeTrafficSplit(backingService, map[string]string{"foo": "bar"}),
				},
				AppliedFederation: &discoveryv1.DestinationStatus_AppliedFederation{
					VirtualMeshRef:    destinationVirtualMeshRef,
					FederatedHostname: "federated-hostname",
					FederatedToMeshes: []*skv2corev1.ObjectRef{
						remoteMeshRef,
						remoteMeshRef2,
					},
				},
			},
		}

		in := input.NewInputLocalSnapshotManualBuilder("ignored").
			AddDestinations(discoveryv1.DestinationSlice{destination}).
			AddMeshes(discoveryv1.MeshSlice{destinationMesh, remoteMesh, remoteMesh2}).
			AddVirtualMeshes(networkingv1.VirtualMeshSlice{destinationVirtualMesh}).
			Build()

		expectedVS := &networkingv1alpha3.VirtualService{}
		mockVirtualServiceTranslator.
			EXPECT().
			Translate(ctx, in, destination, remoteMesh.Spec.GetIstio().Installation, mockReporter).
			Return(expectedVS)
		mockVirtualServiceTranslator.
			EXPECT().
			Translate(ctx, in, destination, remoteMesh2.Spec.GetIstio().Installation, mockReporter).
			Return(expectedVS)

		expectedDR := &networkingv1alpha3.DestinationRule{}
		mockDestinationRuleTranslator.
			EXPECT().
			Translate(ctx, in, destination, remoteMesh.Spec.GetIstio().Installation, mockReporter).
			Return(expectedDR)
		mockDestinationRuleTranslator.
			EXPECT().
			Translate(ctx, in, destination, remoteMesh2.Spec.GetIstio().Installation, mockReporter).
			Return(expectedDR)

		expectedServiceEntry := &networkingv1alpha3.ServiceEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "federated-hostname",
				Namespace:   "remote-namespace",
				ClusterName: "remote-cluster",
				Labels:      metautils.TranslatedObjectLabels(),
				Annotations: map[string]string{
					metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"virtual-mesh","namespace":"namespace"}]}`,
				},
			},
			Spec: networkingv1alpha3spec.ServiceEntry{
				Hosts: []string{
					"federated-hostname",
				},
				Addresses: []string{
					"243.21.204.125",
				},
				Ports: []*networkingv1alpha3spec.Port{
					{
						Number:   1234,
						Protocol: string(protocol.HTTP),
						Name:     "http",
					},
				},
				Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
				Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
				Endpoints: []*networkingv1alpha3spec.WorkloadEntry{
					{
						Address: "mesh-gateway.dns.name",
						Ports: map[string]uint32{
							"http": 8181,
						},
						Labels: map[string]string{"cluster": "cluster"},
					},
				},
			},
		}

		expectedServiceEntry2 := expectedServiceEntry.DeepCopy()
		expectedServiceEntry.Namespace = "remote-namespace2"
		expectedServiceEntry.ClusterName = "remote-cluster2"

		expectedServiceEntries := []*networkingv1alpha3.ServiceEntry{
			expectedServiceEntry,
			expectedServiceEntry2,
		}

		serviceEntries, virtualServices, destinationRules := federationTranslator.Translate(in, destination, mockReporter)

		Expect(serviceEntries).To(ConsistOf(expectedServiceEntries))
		Expect(virtualServices).To(ConsistOf([]*networkingv1alpha3.VirtualService{expectedVS, expectedVS}))
		Expect(destinationRules).To(ConsistOf([]*networkingv1alpha3.DestinationRule{expectedDR, expectedDR}))
	})
})
