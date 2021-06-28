package federation_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/istio"
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
				}},
			},
			Status: discoveryv1.MeshStatus{
				EastWestIngressGateways: []*discoveryv1.MeshStatus_IngressGateway{
					&discoveryv1.MeshStatus_IngressGateway{
						DestinationRef: &skv2corev1.ObjectRef{
							Name:      "istio-ingressgateway",
							Namespace: "istio-system",
						},
						TlsPortName: "tls2",
					},
				},
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

		federatedMeshIngressGatewayDestination := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "istio-system",
				Name:      "istio-ingressgateway",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							Name:        "istio-ingressgateway",
							Namespace:   "istio-system",
							ClusterName: clusterName,
						},
						WorkloadSelectorLabels: map[string]string{"gatewaylabels": "righthere"},
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
							&discoveryv1.DestinationSpec_KubeService_KubeServicePort{
								Port:     80,
								Name:     "http2",
								Protocol: "TCP",
								TargetPort: &discoveryv1.DestinationSpec_KubeService_KubeServicePort_TargetPortNumber{
									TargetPortNumber: 8080,
								},
								NodePort: 30742,
							},
							&discoveryv1.DestinationSpec_KubeService_KubeServicePort{
								Port:     9191,
								Name:     "tls2",
								Protocol: "TCP",
								TargetPort: &discoveryv1.DestinationSpec_KubeService_KubeServicePort_TargetPortNumber{
									TargetPortNumber: 15443,
								},
								NodePort: 32001,
							},
						},
						ServiceType: discoveryv1.DestinationSpec_KubeService_NODE_PORT,
					},
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
					EastWestIngressGatewaySelectors: []*commonv1.IngressGatewaySelector{
						&commonv1.IngressGatewaySelector{
							Meshes: []*skv2corev1.ObjectRef{
								ezkube.MakeObjectRef(mesh),
							},
							DestinationSelectors: []*commonv1.DestinationSelector{
								{
									KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
										Services: []*skv2corev1.ClusterObjectRef{
											{
												Name:        "istio-ingressgateway",
												Namespace:   "istio-system",
												ClusterName: clusterName,
											},
										},
									},
								},
							},
							GatewayTlsPortName: "tls2",
						},
					},
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
			AddMeshes(discoveryv1.MeshSlice{mesh, clientMesh}).
			AddKubernetesClusters(skv1alpha1.KubernetesClusterSlice{kubeCluster}).
			AddDestinations(discoveryv1.DestinationSlice{federatedMeshIngressGatewayDestination}).
			Build()

		t := NewTranslator(ctx)

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
	})
})

var (
	expectedGateway = &networkingv1alpha3.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-virtual-mesh-config-namespace-istio-ingressgateway",
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
)
