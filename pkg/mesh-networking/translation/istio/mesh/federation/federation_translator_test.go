package federation_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
						DestinationRef: &skv2corev1.ObjectRef{
							Name:      "istio-ingressgateway",
							Namespace: "istio-system",
						},
						DestinationPort: 1234,
						ContainerPort:   91234,
					},
					{
						DestinationRef: &skv2corev1.ObjectRef{
							Name:      "istio-ingressgateway2",
							Namespace: "istio-system",
						},
						DestinationPort: 5678,
						ContainerPort:   95678,
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
					Installation: &discoveryv1.MeshInstallation{
						Namespace: "remote-namespace",
						Cluster:   "remote-cluster",
					},
				}},
			},
		}

		meshRef := ezkube.MakeObjectRef(mesh)
		clientMeshRef := ezkube.MakeObjectRef(clientMesh)

		destination1 := &discoveryv1.Destination{
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
					},
				},
			},
		}
		destination2 := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "istio-system",
				Name:      "istio-ingressgateway2",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							Name:        "istio-ingressgateway2",
							Namespace:   "istio-system",
							ClusterName: clusterName,
						},
						WorkloadSelectorLabels: map[string]string{"foo": "bar"},
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
			AddDestinations(discoveryv1.DestinationSlice{destination1, destination2}).
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

		expectedGateways := []*networkingv1alpha3.Gateway{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "istio-ingressgateway-istio-system",
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
								Number:   91234,
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
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "istio-ingressgateway2-istio-system",
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
								Number:   95678,
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
					Selector: map[string]string{"foo": "bar"},
				},
			},
		}

		Expect(outputs.GetGateways().List()).To(ConsistOf(expectedGateways))
	})
})
