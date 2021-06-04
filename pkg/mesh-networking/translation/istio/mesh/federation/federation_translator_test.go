package federation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
	mock_destinationrule "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/destinationrule/mocks"
	mock_virtualservice "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/destination/virtualservice/mocks"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/gloo-mesh/test/data"
	"istio.io/istio/pkg/config/protocol"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istiov1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	. "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/istio/mesh/federation"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	skv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	"github.com/solo-io/skv2/pkg/ezkube"
	networkingv1alpha3spec "istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FederationTranslator", func() {
	ctx := context.TODO()
	ctrl := gomock.NewController(GinkgoT())
	mockVirtualServiceTranslator := mock_virtualservice.NewMockTranslator(ctrl)
	mockDestinationRuleTranslator := mock_destinationrule.NewMockTranslator(ctrl)

	It("translates federation resources for a VirtualMesh", func() {

		namespace := "namespace"
		clusterName := "cluster"

		mesh := &discoveryv1.Mesh{
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

		clientMesh := &discoveryv1.Mesh{
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

		meshRef := ezkube.MakeObjectRef(mesh)
		clientMeshRef := ezkube.MakeObjectRef(clientMesh)

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
		destination1 := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: backingService,
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
							{
								Port:     1234,
								Name:     "http",
								Protocol: "TCP",
							},
							{
								Port:     5555,
								Protocol: "TCP", // translated ServiceEntry should fall back on protocol for port name because name isn't specified here
							},
							{
								Port:     5678,
								Name:     "grpc",
								Protocol: "TCP",
							},
						},
						EndpointSubsets: []*discoveryv1.DestinationSpec_KubeService_EndpointsSubset{
							{
								Endpoints: []*discoveryv1.DestinationSpec_KubeService_EndpointsSubset_Endpoint{
									{
										IpAddress: "192.168.21.1",
										Labels: map[string]string{
											"version": "v1",
										},
									},
									{
										IpAddress: "192.168.21.2",
										Labels: map[string]string{
											"version": "v2",
										},
									},
								},
								Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
									{
										Port:     1234,
										Protocol: "TCP",
									},
								},
							},
						},
					},
				},
				Mesh: meshRef,
			},
			// include some applied subsets
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*discoveryv1.DestinationStatus_AppliedTrafficPolicy{
					makeTrafficSplit(backingService, map[string]string{"foo": "bar"}),
					makeTrafficSplit(backingService, map[string]string{"foo": "baz"}),
					makeTrafficSplit(backingService, map[string]string{"bar": "qux"}),
				},
			},
		}

		vMesh := &discoveryv1.MeshStatus_AppliedVirtualMesh{
			Ref: &skv2corev1.ObjectRef{
				Name:      "my-virtual-mesh",
				Namespace: "config-namespace",
			},
			Spec: &v1.VirtualMeshSpec{
				Meshes: []*skv2corev1.ObjectRef{
					meshRef,
					clientMeshRef,
				},
				Federation: &v1.VirtualMeshSpec_Federation{
					HostnameSuffix: "soloio",
				},
			},
		}

		kubeCluster := &skv1alpha1.KubernetesCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: defaults.GetPodNamespace(),
			},
		}

		in := input.NewInputLocalSnapshotManualBuilder("ignored").
			AddDestinations(discoveryv1.DestinationSlice{destination1}).
			AddMeshes(discoveryv1.MeshSlice{mesh, clientMesh}).
			AddKubernetesClusters(skv1alpha1.KubernetesClusterSlice{kubeCluster}).
			Build()

		expectedRemoteVS := &networkingv1alpha3.VirtualService{}
		mockVirtualServiceTranslator.
			EXPECT().
			Translate(ctx, in, destination1, clientMesh.Spec.GetIstio().Installation, nil).
			Return(expectedRemoteVS)

		expectedRemoteDR := &networkingv1alpha3.DestinationRule{
			Spec: networkingv1alpha3spec.DestinationRule{
				Subsets: []*networkingv1alpha3spec.Subset{
					{
						Name: "version-v1",
						Labels: map[string]string{
							"version": "v1",
						},
					},
					{
						Name: "version-v2",
						Labels: map[string]string{
							"version": "v2",
						},
					},
				},
				TrafficPolicy: &networkingv1alpha3spec.TrafficPolicy{
					Tls: &networkingv1alpha3spec.ClientTLSSettings{
						Mode: networkingv1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
					},
				},
			},
		}
		mockDestinationRuleTranslator.
			EXPECT().
			Translate(ctx, in, destination1, clientMesh.Spec.GetIstio().Installation, nil).
			Return(expectedRemoteDR)

		expectedRemoteServiceEntry := &networkingv1alpha3.ServiceEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "some-svc.some-ns.svc.cluster.soloio",
				Namespace:   "remote-namespace",
				ClusterName: "remote-cluster",
				Labels:      metautils.TranslatedObjectLabels(),
				Annotations: map[string]string{
					metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"my-virtual-mesh","namespace":"config-namespace"}]}`,
				},
			},
			Spec: networkingv1alpha3spec.ServiceEntry{
				Hosts: []string{
					"some-svc.some-ns.svc.cluster.soloio",
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
					{
						Number:   5555,
						Protocol: string(protocol.TCP),
						Name:     "TCP",
					},
					{
						Number:   5678,
						Protocol: string(protocol.GRPC),
						Name:     "grpc",
					},
				},
				Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
				Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
				Endpoints: []*networkingv1alpha3spec.WorkloadEntry{
					{
						Address: "mesh-gateway.dns.name",
						Ports: map[string]uint32{
							"TCP":  8181,
							"http": 8181,
							"grpc": 8181,
						},
						Labels: map[string]string{"version": "v1"},
					},
					{
						Address: "mesh-gateway.dns.name",
						Ports: map[string]uint32{
							"TCP":  8181,
							"http": 8181,
							"grpc": 8181,
						},
						Labels: map[string]string{"version": "v2"},
					},
				},
			},
		}

		federatedHostname := hostutils.BuildFederatedFQDN(
			destination1.Spec.GetKubeService().Ref,
			vMesh.Spec,
		)
		expectedLocalServiceEntry := &networkingv1alpha3.ServiceEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:        federatedHostname,
				Namespace:   mesh.Spec.GetIstio().Installation.Namespace,
				ClusterName: mesh.Spec.GetIstio().Installation.Cluster,
				Labels:      metautils.TranslatedObjectLabels(),
				Annotations: map[string]string{
					metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"my-virtual-mesh","namespace":"config-namespace"}]}`,
				},
			},
			Spec: networkingv1alpha3spec.ServiceEntry{
				// match the federate hostname
				Hosts: []string{federatedHostname},
				// only export to Gateway workload namespace
				ExportTo:   []string{"."},
				Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
				Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
				Endpoints: []*networkingv1alpha3spec.WorkloadEntry{
					{
						// map to the local hostname
						Address: "192.168.21.1",
						// needed for cross cluster subset routing
						Labels: map[string]string{"version": "v1"},
						Ports:  map[string]uint32{"TCP": 1234},
					},
					{
						// map to the local hostname
						Address: "192.168.21.2",
						// needed for cross cluster subset routing
						Labels: map[string]string{"version": "v2"},
						Ports:  map[string]uint32{"TCP": 1234},
					},
				},
				Ports: expectedRemoteServiceEntry.Spec.Ports,
			},
		}

		expectedLocalDestinationRule := &networkingv1alpha3.DestinationRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:        federatedHostname,
				Namespace:   mesh.Spec.GetIstio().Installation.Namespace,
				ClusterName: mesh.Spec.GetIstio().Installation.Cluster,
				Labels:      metautils.TranslatedObjectLabels(),
				Annotations: map[string]string{
					metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"my-virtual-mesh","namespace":"config-namespace"}]}`,
				},
			},
			Spec: networkingv1alpha3spec.DestinationRule{
				Host:          federatedHostname,
				Subsets:       expectedRemoteDR.Spec.Subsets,
				TrafficPolicy: expectedRemoteDR.Spec.TrafficPolicy,
			},
		}

		t := NewTranslator(
			ctx,
			in.Destinations(),
			mockVirtualServiceTranslator,
			mockDestinationRuleTranslator,
		)

		outputs := istio.NewBuilder(context.TODO(), "")
		t.Translate(
			in,
			mesh,
			vMesh,
			outputs,
			nil, // no reports expected
		)

		Expect(outputs.GetGateways().Length()).To(Equal(1))
		Expect(outputs.GetGateways().List()[0]).To(Equal(expectedGateway))
		Expect(outputs.GetDestinationRules().List()).To(ConsistOf([]*networkingv1alpha3.DestinationRule{expectedRemoteDR, expectedLocalDestinationRule}))
		Expect(outputs.GetServiceEntries().List()).To(ConsistOf([]*networkingv1alpha3.ServiceEntry{expectedRemoteServiceEntry, expectedLocalServiceEntry}))
		Expect(outputs.GetVirtualServices()).To(Equal(istiov1alpha3sets.NewVirtualServiceSet(expectedRemoteVS)))
	})
})

var expectedGateway = &networkingv1alpha3.Gateway{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "my-virtual-mesh-config-namespace",
		Namespace:   "namespace",
		ClusterName: "cluster",
		Labels:      metautils.TranslatedObjectLabels(),
		Annotations: map[string]string{
			metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"my-virtual-mesh","namespace":"config-namespace"}]}`,
		},
	},
	Spec: networkingv1alpha3spec.Gateway{
		Servers: []*networkingv1alpha3spec.Server{
			{
				Port: &networkingv1alpha3spec.Port{
					Number:   9191,
					Protocol: "TLS",
					Name:     "tls",
				},
				Hosts: []string{
					"*.soloio",
				},
				Tls: &networkingv1alpha3spec.ServerTLSSettings{
					Mode: networkingv1alpha3spec.ServerTLSSettings_AUTO_PASSTHROUGH,
				},
			},
		},
		Selector: map[string]string{"gatewaylabels": "righthere"},
	},
}
