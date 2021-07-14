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
					Installation: &discoveryv1.MeshInstallation{
						Namespace: namespace,
						Cluster:   clusterName,
						Version:   "1.8.1",
					},
				}},
			},
			Status: discoveryv1.MeshStatus{
				AppliedEastWestIngressGateways: []*discoveryv1.MeshStatus_AppliedIngressGateway{
					{
						DestinationRef:    nil, // not used at this stage in translation
						ExternalAddresses: []string{"mesh-gateway.dns.name"},
						DestinationPort:   8181,
					},
					{
						DestinationRef:    nil, // not used at this stage in translation
						ExternalAddresses: []string{"172.18.0.2"},
						DestinationPort:   8181,
					},
				},
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
					Installation: &discoveryv1.MeshInstallation{
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
					Installation: &discoveryv1.MeshInstallation{
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

		makeTrafficSplit := func(backingService *skv2corev1.ClusterObjectRef, subset map[string]string) *networkingv1.AppliedTrafficPolicy {
			return &networkingv1.AppliedTrafficPolicy{Spec: &data.RemoteTrafficShiftPolicy(
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
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: backingService,
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
							{
								Port:     1234,
								Protocol: "TCP", // translated ServiceEntry should fall back on protocol for port name because name isn't specified here
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
								Ports: []*discoveryv1.DestinationSpec_KubeService_EndpointPort{
									{
										Port:     1234,
										Protocol: "TCP",
									},
								},
							},
						},
					},
				},
				Mesh: destinationMeshRef,
			},
			// include some applied subsets
			Status: discoveryv1.DestinationStatus{
				AppliedTrafficPolicies: []*networkingv1.AppliedTrafficPolicy{
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

		expectedRemoteVS := &networkingv1alpha3.VirtualService{}
		mockVirtualServiceTranslator.
			EXPECT().
			Translate(ctx, in, destination, remoteMesh.Spec.GetIstio().Installation, mockReporter).
			Return(expectedRemoteVS)
		mockVirtualServiceTranslator.
			EXPECT().
			Translate(ctx, in, destination, remoteMesh2.Spec.GetIstio().Installation, mockReporter).
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
			Translate(ctx, in, destination, remoteMesh.Spec.GetIstio().Installation, mockReporter).
			Return(expectedRemoteDR)
		mockDestinationRuleTranslator.
			EXPECT().
			Translate(ctx, in, destination, remoteMesh2.Spec.GetIstio().Installation, mockReporter).
			Return(expectedRemoteDR)

		expectedRemoteServiceEntry := &networkingv1alpha3.ServiceEntry{
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
						Protocol: string(protocol.TCP),
						Name:     "TCP",
					},
				},
				Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
				Resolution: networkingv1alpha3spec.ServiceEntry_DNS,
				Endpoints: []*networkingv1alpha3spec.WorkloadEntry{
					{
						Address: "mesh-gateway.dns.name",
						Ports: map[string]uint32{
							"TCP": 8181,
						},
						Labels: map[string]string{"version": "v1"},
					},
					{
						Address: "mesh-gateway.dns.name",
						Ports: map[string]uint32{
							"TCP": 8181,
						},
						Labels: map[string]string{"version": "v2"},
					},
					{
						Address: "172.18.0.2",
						Ports: map[string]uint32{
							"TCP": 8181,
						},
						Labels: map[string]string{"version": "v1"},
					},
					{
						Address: "172.18.0.2",
						Ports: map[string]uint32{
							"TCP": 8181,
						},
						Labels: map[string]string{"version": "v2"},
					},
				},
			},
		}

		expectedRemoteServiceEntry2 := expectedRemoteServiceEntry.DeepCopy()

		expectedRemoteServiceEntry.Namespace = "remote-namespace2"
		expectedRemoteServiceEntry.ClusterName = "remote-cluster2"

		expectedLocalServiceEntry := &networkingv1alpha3.ServiceEntry{
			ObjectMeta: metav1.ObjectMeta{
				Name:        destination.Status.GetAppliedFederation().FederatedHostname,
				Namespace:   destinationMesh.Spec.GetIstio().Installation.Namespace,
				ClusterName: destinationMesh.Spec.GetIstio().Installation.Cluster,
				Labels:      metautils.TranslatedObjectLabels(),
				Annotations: map[string]string{
					metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"virtual-mesh","namespace":"namespace"}]}`,
				},
			},
			Spec: networkingv1alpha3spec.ServiceEntry{
				// match the federate hostname
				Hosts: []string{destination.Status.GetAppliedFederation().FederatedHostname},
				// only export to Gateway workload namespace
				ExportTo:   []string{"."},
				Location:   networkingv1alpha3spec.ServiceEntry_MESH_INTERNAL,
				Resolution: networkingv1alpha3spec.ServiceEntry_STATIC,
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
				Name:        destination.Status.GetAppliedFederation().FederatedHostname,
				Namespace:   destinationMesh.Spec.GetIstio().Installation.Namespace,
				ClusterName: destinationMesh.Spec.GetIstio().Installation.Cluster,
				Labels:      metautils.TranslatedObjectLabels(),
				Annotations: map[string]string{
					metautils.ParentLabelkey: `{"networking.mesh.gloo.solo.io/v1, Kind=VirtualMesh":[{"name":"virtual-mesh","namespace":"namespace"}]}`,
				},
			},
			Spec: networkingv1alpha3spec.DestinationRule{
				Host:          destination.Status.GetAppliedFederation().FederatedHostname,
				Subsets:       expectedRemoteDR.Spec.Subsets,
				TrafficPolicy: expectedRemoteDR.Spec.TrafficPolicy,
			},
		}

		serviceEntries, virtualServices, destinationRules := federationTranslator.Translate(in, destination, mockReporter)

		Expect(serviceEntries).To(ConsistOf([]*networkingv1alpha3.ServiceEntry{expectedRemoteServiceEntry, expectedRemoteServiceEntry2, expectedLocalServiceEntry}))
		Expect(virtualServices).To(ConsistOf([]*networkingv1alpha3.VirtualService{expectedRemoteVS, expectedRemoteVS}))
		Expect(destinationRules).To(ConsistOf([]*networkingv1alpha3.DestinationRule{expectedRemoteDR, expectedRemoteDR, expectedLocalDestinationRule}))
	})

	It("should set ServiceEntry resolution to STATIC if any endpoints have ipv6 address", func() {
		workloadEntries := []*networkingv1alpha3spec.WorkloadEntry{
			{
				Address: "192.168.21.2",
			},
		}
		res, err := federation.ResolutionForEndpointIpVersions(workloadEntries)
		Expect(err).To(BeNil())
		Expect(res).To(Equal(networkingv1alpha3spec.ServiceEntry_STATIC))

		workloadEntriesWithIpv6 := append(workloadEntries, &networkingv1alpha3spec.WorkloadEntry{Address: "fd00:10:96::dcc0"})
		res, err = federation.ResolutionForEndpointIpVersions(workloadEntriesWithIpv6)
		Expect(err).To(BeNil())
		Expect(res).To(Equal(networkingv1alpha3spec.ServiceEntry_STATIC))

		// ipv4 addresses represented as ipv6 should still be handled with STATIC resolution
		workloadEntriesWithIpv4MappedIpv6 := append(workloadEntries, &networkingv1alpha3spec.WorkloadEntry{Address: "::FFFF:C0A8:1"})
		res, err = federation.ResolutionForEndpointIpVersions(workloadEntriesWithIpv4MappedIpv6)
		Expect(err).To(BeNil())
		Expect(res).To(Equal(networkingv1alpha3spec.ServiceEntry_STATIC))

		// ipv4 addresses represented as ipv6 should still be handled with STATIC resolution
		workloadEntriesWithIpv4MappedIpv6 = append(workloadEntries, &networkingv1alpha3spec.WorkloadEntry{Address: "::FFFF:C0A8:0001"})
		res, err = federation.ResolutionForEndpointIpVersions(workloadEntriesWithIpv4MappedIpv6)
		Expect(err).To(BeNil())
		Expect(res).To(Equal(networkingv1alpha3spec.ServiceEntry_STATIC))

		// error should be thrown if both ipv6 and hostname found
		workloadEntriesWithIpv4MappedIpv6 = append(workloadEntries, &networkingv1alpha3spec.WorkloadEntry{Address: "::FFFF:C0A8:0001"})
		workloadEntriesWithIpv4MappedIpv6 = append(workloadEntriesWithIpv4MappedIpv6, &networkingv1alpha3spec.WorkloadEntry{Address: "hostname"})
		_, err = federation.ResolutionForEndpointIpVersions(workloadEntriesWithIpv4MappedIpv6)
		Expect(err).ToNot(BeNil())

		workloadEntriesWithHostname := append(workloadEntries, &networkingv1alpha3spec.WorkloadEntry{Address: "hostname"})
		res, err = federation.ResolutionForEndpointIpVersions(workloadEntriesWithHostname)
		Expect(err).To(BeNil())
		Expect(res).To(Equal(networkingv1alpha3spec.ServiceEntry_DNS))
	})
})
