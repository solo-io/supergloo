package federation_test

import (
	"context"

	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	istiov1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discoveryv1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	skv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/federation"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/hostutils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	istiov1alpha3spec "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("FederationTranslator", func() {
	ctx := context.TODO()
	clusterDomains := hostutils.NewClusterDomainRegistry(skv1alpha1sets.NewKubernetesClusterSet())

	It("translates federation resources for a virtual mesh", func() {

		namespace := "namespace"
		clusterName := "cluster"

		mesh := &discoveryv1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "config-namespace",
				Name:      "federated-mesh",
			},
			Spec: discoveryv1alpha1.MeshSpec{
				MeshType: &discoveryv1alpha1.MeshSpec_Istio_{Istio: &discoveryv1alpha1.MeshSpec_Istio{
					Installation: &discoveryv1alpha1.MeshSpec_MeshInstallation{
						Namespace: namespace,
						Cluster:   clusterName,
					},
					IngressGateways: []*discoveryv1alpha1.MeshSpec_Istio_IngressGatewayInfo{{
						ExternalAddress:  "mesh-gateway.dns.name",
						ExternalTlsPort:  8181,
						TlsContainerPort: 9191,
						WorkloadLabels:   map[string]string{"gatewaylabels": "righthere"},
					}},
				}},
			},
		}

		clientMesh := &discoveryv1alpha1.Mesh{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "config-namespace",
				Name:      "client-mesh",
			},
			Spec: discoveryv1alpha1.MeshSpec{
				MeshType: &discoveryv1alpha1.MeshSpec_Istio_{Istio: &discoveryv1alpha1.MeshSpec_Istio{
					Installation: &discoveryv1alpha1.MeshSpec_MeshInstallation{
						Namespace: "remote-namespace",
						Cluster:   "remote-cluster",
					},
				}},
			},
		}

		meshRef := ezkube.MakeObjectRef(mesh)
		clientMeshRef := ezkube.MakeObjectRef(clientMesh)

		meshService1 := &discoveryv1alpha1.MeshService{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: discoveryv1alpha1.MeshServiceSpec{
				Type: &discoveryv1alpha1.MeshServiceSpec_KubeService_{KubeService: &discoveryv1alpha1.MeshServiceSpec_KubeService{
					Ref: &v1.ClusterObjectRef{
						Name:        "some-svc",
						Namespace:   "some-ns",
						ClusterName: clusterName,
					},
					Ports: []*discoveryv1alpha1.MeshServiceSpec_KubeService_KubeServicePort{
						{
							Port:     1234,
							Name:     "http",
							Protocol: "TCP",
						},
						{
							Port:     5678,
							Name:     "grpc",
							Protocol: "TCP",
						},
					},
				}},
				Mesh: meshRef,
			},
		}

		vMesh := &discoveryv1alpha1.MeshStatus_AppliedVirtualMesh{
			Ref: &v1.ObjectRef{
				Name:      "my-virtual-mesh",
				Namespace: "config-namespace",
			},
			Spec: &v1alpha1.VirtualMeshSpec{
				Meshes: []*v1.ObjectRef{
					meshRef,
					clientMeshRef,
				},
			},
		}

		kubeCluster := &skv1alpha1.KubernetesCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterName,
				Namespace: defaults.GetPodNamespace(),
			},
		}

		in := input.NewSnapshot(
			"ignored",
			discoveryv1alpha1sets.NewMeshServiceSet(meshService1), discoveryv1alpha1sets.NewMeshWorkloadSet(), discoveryv1alpha1sets.NewMeshSet(mesh, clientMesh),

			v1alpha1sets.NewTrafficPolicySet(),
			v1alpha1sets.NewAccessPolicySet(),
			v1alpha1sets.NewVirtualMeshSet(),

			skv1alpha1sets.NewKubernetesClusterSet(kubeCluster),
		)

		t := NewTranslator(ctx, clusterDomains)
		outputs := t.Translate(
			in,
			mesh,
			vMesh,
			nil, // no reports expected
		)

		Expect(outputs).To(Equal(expected))

	})
})

var expected = Outputs{
	Gateway: &istiov1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-virtual-mesh.config-namespace",
			Namespace:   "namespace",
			ClusterName: "cluster",
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: istiov1alpha3spec.Gateway{
			Servers: []*istiov1alpha3spec.Server{
				{
					Port: &istiov1alpha3spec.Port{
						Number:   9191,
						Protocol: "TLS",
						Name:     "tls",
					},
					Hosts: []string{
						"some-svc.some-ns.svc.cluster",
					},
					Tls: &istiov1alpha3spec.ServerTLSSettings{
						Mode: istiov1alpha3spec.ServerTLSSettings_AUTO_PASSTHROUGH,
					},
				},
			},
			Selector: map[string]string{"gatewaylabels": "righthere"},
		},
	},
	EnvoyFilter: &istiov1alpha3.EnvoyFilter{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-virtual-mesh.config-namespace",
			Namespace:   "namespace",
			ClusterName: "cluster",
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: istiov1alpha3spec.EnvoyFilter{
			WorkloadSelector: &istiov1alpha3spec.WorkloadSelector{
				Labels: map[string]string{"gatewaylabels": "righthere"},
			},
			ConfigPatches: []*istiov1alpha3spec.EnvoyFilter_EnvoyConfigObjectPatch{
				{
					ApplyTo: istiov1alpha3spec.EnvoyFilter_NETWORK_FILTER,
					Match: &istiov1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch{
						Context: istiov1alpha3spec.EnvoyFilter_GATEWAY,
						ObjectTypes: &istiov1alpha3spec.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
							Listener: &istiov1alpha3spec.EnvoyFilter_ListenerMatch{
								PortNumber: 9191,
								FilterChain: &istiov1alpha3spec.EnvoyFilter_ListenerMatch_FilterChainMatch{
									Filter: &istiov1alpha3spec.EnvoyFilter_ListenerMatch_FilterMatch{
										Name: "envoy.filters.network.sni_cluster",
									},
								},
							},
						},
					},
					Patch: &istiov1alpha3spec.EnvoyFilter_Patch{
						Operation: 5,
						Value: &types.Struct{
							Fields: map[string]*types.Value{
								"name": {
									Kind: &types.Value_StringValue{
										StringValue: "envoy.filters.network.tcp_cluster_rewrite",
									},
								},
								"config": {
									Kind: &types.Value_StructValue{
										StructValue: &types.Struct{
											Fields: map[string]*types.Value{
												"cluster_replacement": {
													Kind: &types.Value_StringValue{
														StringValue: ".svc.cluster.local",
													},
												},
												"cluster_pattern": {
													Kind: &types.Value_StringValue{
														StringValue: "\\.cluster$",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
	DestinationRules: istiov1alpha3sets.NewDestinationRuleSet(&istiov1alpha3.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "some-svc.some-ns.svc.cluster",
			Namespace:   "remote-namespace",
			ClusterName: "remote-cluster",
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: istiov1alpha3spec.DestinationRule{
			Host: "some-svc.some-ns.svc.cluster",
			TrafficPolicy: &istiov1alpha3spec.TrafficPolicy{
				Tls: &istiov1alpha3spec.ClientTLSSettings{
					Mode: istiov1alpha3spec.ClientTLSSettings_ISTIO_MUTUAL,
				},
			},
		},
	}),
	ServiceEntries: istiov1alpha3sets.NewServiceEntrySet(&istiov1alpha3.ServiceEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "some-svc.some-ns.svc.cluster",
			Namespace:   "remote-namespace",
			ClusterName: "remote-cluster",
			Labels:      metautils.TranslatedObjectLabels(),
		},
		Spec: istiov1alpha3spec.ServiceEntry{
			Hosts: []string{
				"some-svc.some-ns.svc.cluster",
			},
			Addresses: []string{
				"242.147.203.114",
			},
			Ports: []*istiov1alpha3spec.Port{
				{
					Number:   1234,
					Protocol: "TCP",
					Name:     "http",
				},
				{
					Number:   5678,
					Protocol: "TCP",
					Name:     "grpc",
				},
			},
			Location:   istiov1alpha3spec.ServiceEntry_MESH_INTERNAL,
			Resolution: istiov1alpha3spec.ServiceEntry_DNS,
			Endpoints: []*istiov1alpha3spec.WorkloadEntry{
				{
					Address: "mesh-gateway.dns.name",
					Ports: map[string]uint32{
						"http": 8181,
						"grpc": 8181,
					},
				},
			},
		},
	}),
}
