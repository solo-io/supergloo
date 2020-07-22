package validation_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	networkingv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/mesh/failoverservice/validation"
	corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	multiclusterv1alpha1 "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
	multiclusterv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validation", func() {
	var (
		ctrl      *gomock.Controller
		validator validation.FailoverServiceValidator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		validator = validation.NewFailoverServiceValidator()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var meshServiceStatusWithOutlierDetection = func() discoveryv1alpha2.MeshServiceStatus {
		return discoveryv1alpha2.MeshServiceStatus{
			AppliedTrafficPolicies: []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
				{

					Spec: &networkingv1alpha2.TrafficPolicySpec{
						OutlierDetection: &networkingv1alpha2.TrafficPolicySpec_OutlierDetection{},
					},
				},
			},
		}
	}

	// Snapshot with valid FailoverService.
	var validInputs = func() (validation.Inputs, *networkingv1alpha2.FailoverServiceSpec) {
		return validation.Inputs{
				MeshServices: discoveryv1alpha2sets.NewMeshServiceSet(
					&discoveryv1alpha2.MeshService{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "meshservice1",
							Namespace:   "namespace1",
							ClusterName: "cluster2",
						},
						Spec: discoveryv1alpha2.MeshServiceSpec{
							Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
								KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
									Ref: &corev1.ClusterObjectRef{
										Name:        "service1",
										Namespace:   "namespace1",
										ClusterName: "cluster2",
									},
								},
							},
							Mesh: &corev1.ObjectRef{
								Name:      "mesh2",
								Namespace: "namespace1",
							},
						},
						Status: meshServiceStatusWithOutlierDetection(),
					},
				),
				KubeClusters: multiclusterv1alpha1sets.NewKubernetesClusterSet(
					&multiclusterv1alpha1.KubernetesCluster{ObjectMeta: metav1.ObjectMeta{Name: "cluster1"}},
					&multiclusterv1alpha1.KubernetesCluster{ObjectMeta: metav1.ObjectMeta{Name: "cluster2"}},
				),
				Meshes: discoveryv1alpha2sets.NewMeshSet(
					&discoveryv1alpha2.Mesh{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mesh1",
							Namespace: "namespace1",
						},
						Spec: discoveryv1alpha2.MeshSpec{
							MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
								Istio: &discoveryv1alpha2.MeshSpec_Istio{
									Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
										Cluster: "cluster1",
									},
								},
							},
						},
						Status: discoveryv1alpha2.MeshStatus{
							AppliedVirtualMeshes: []*discoveryv1alpha2.MeshStatus_AppliedVirtualMesh{
								{
									Ref: &corev1.ObjectRef{
										Name:      "virtual-mesh1",
										Namespace: "namespace1",
									},
								},
							},
						},
					},
					&discoveryv1alpha2.Mesh{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mesh2",
							Namespace: "namespace1",
						},
						Spec: discoveryv1alpha2.MeshSpec{
							MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
								Istio: &discoveryv1alpha2.MeshSpec_Istio{
									Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
										Cluster: "cluster2",
									},
								},
							},
						},
						Status: discoveryv1alpha2.MeshStatus{
							AppliedVirtualMeshes: []*discoveryv1alpha2.MeshStatus_AppliedVirtualMesh{
								{
									Ref: &corev1.ObjectRef{
										Name:      "virtual-mesh1",
										Namespace: "namespace1",
									},
								},
							},
						},
					}),
				VirtualMeshes: networkingv1alpha2sets.NewVirtualMeshSet(
					&networkingv1alpha2.VirtualMesh{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "virtual-mesh1",
							Namespace: "namespace1",
						},
						Spec: networkingv1alpha2.VirtualMeshSpec{
							Meshes: []*corev1.ObjectRef{
								{
									Name:      "mesh1",
									Namespace: "namespace1",
								},
								{
									Name:      "mesh2",
									Namespace: "namespace1",
								},
							},
						},
					},
				),
			}, &networkingv1alpha2.FailoverServiceSpec{
				Hostname: "service1.namespace1.cluster1",
				Port: &networkingv1alpha2.FailoverServiceSpec_Port{
					Port:     9080,
					Protocol: "http",
				},
				Meshes: []*corev1.ObjectRef{
					{
						Name:      "mesh1",
						Namespace: "namespace1",
					},
				},
				FailoverServices: []*corev1.ClusterObjectRef{
					{
						Name:        "service1",
						Namespace:   "namespace1",
						ClusterName: "cluster2",
					},
				},
			}
	}

	// Snapshot with valid FailoverService in a single mesh.
	var validInputSnapshotSingleMesh = func() (validation.Inputs, *networkingv1alpha2.FailoverServiceSpec) {
		return validation.Inputs{
				MeshServices: discoveryv1alpha2sets.NewMeshServiceSet(
					&discoveryv1alpha2.MeshService{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "meshservice1",
							Namespace:   "namespace1",
							ClusterName: "cluster1",
						},
						Spec: discoveryv1alpha2.MeshServiceSpec{
							Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
								KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
									Ref: &corev1.ClusterObjectRef{
										Name:        "service1",
										Namespace:   "namespace1",
										ClusterName: "cluster1",
									},
								},
							},
							Mesh: &corev1.ObjectRef{
								Name:      "mesh1",
								Namespace: "namespace1",
							},
						},
						Status: meshServiceStatusWithOutlierDetection(),
					},
					&discoveryv1alpha2.MeshService{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "meshservice1",
							Namespace:   "namespace1",
							ClusterName: "cluster2",
						},
						Spec: discoveryv1alpha2.MeshServiceSpec{
							Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
								KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
									Ref: &corev1.ClusterObjectRef{
										Name:        "service1",
										Namespace:   "namespace1",
										ClusterName: "cluster2",
									},
								},
							},
							Mesh: &corev1.ObjectRef{
								Name:      "mesh1",
								Namespace: "namespace1",
							},
						},
						Status: meshServiceStatusWithOutlierDetection(),
					},
				),
				KubeClusters: multiclusterv1alpha1sets.NewKubernetesClusterSet(
					&multiclusterv1alpha1.KubernetesCluster{ObjectMeta: metav1.ObjectMeta{Name: "cluster1"}},
					&multiclusterv1alpha1.KubernetesCluster{ObjectMeta: metav1.ObjectMeta{Name: "cluster2"}},
				),
				Meshes: discoveryv1alpha2sets.NewMeshSet(
					&discoveryv1alpha2.Mesh{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mesh1",
							Namespace: "namespace1",
						},
						Spec: discoveryv1alpha2.MeshSpec{
							MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
								Istio: &discoveryv1alpha2.MeshSpec_Istio{
									Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
										Cluster: "cluster1",
									},
								},
							},
						},
					},
				),
				VirtualMeshes: networkingv1alpha2sets.NewVirtualMeshSet(),
			}, &networkingv1alpha2.FailoverServiceSpec{
				Hostname: "service1.namespace1.cluster1",
				Port: &networkingv1alpha2.FailoverServiceSpec_Port{
					Port:     9080,
					Protocol: "http",
				},
				Meshes: []*corev1.ObjectRef{
					{
						Name:      "mesh1",
						Namespace: "namespace1",
					},
				},
				FailoverServices: []*corev1.ClusterObjectRef{
					{
						Name:        "service1",
						Namespace:   "namespace1",
						ClusterName: "cluster2",
					},
				},
			}
	}

	// Snapshot with invalid FailoverService.
	var invalidInputSnapshot = func() (validation.Inputs, *networkingv1alpha2.FailoverServiceSpec) {
		return validation.Inputs{
				MeshServices: discoveryv1alpha2sets.NewMeshServiceSet(
					&discoveryv1alpha2.MeshService{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "meshservice1",
							Namespace:   "namespace1",
							ClusterName: "cluster1",
						},
						Spec: discoveryv1alpha2.MeshServiceSpec{
							Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
								KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
									Ref: &corev1.ClusterObjectRef{
										Name:        "service1",
										Namespace:   "namespace1",
										ClusterName: "cluster1",
									},
								},
							},
							Mesh: &corev1.ObjectRef{
								Name:      "mesh1",
								Namespace: "namespace1",
							},
						},
						Status: meshServiceStatusWithOutlierDetection(),
					},
					&discoveryv1alpha2.MeshService{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "meshservice1",
							Namespace:   "namespace1",
							ClusterName: "cluster2",
						},
						Spec: discoveryv1alpha2.MeshServiceSpec{
							Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
								KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
									Ref: &corev1.ClusterObjectRef{
										Name:        "service1",
										Namespace:   "namespace1",
										ClusterName: "cluster2",
									},
								},
							},
							Mesh: &corev1.ObjectRef{
								Name:      "mesh2",
								Namespace: "namespace1",
							},
						},
						Status: meshServiceStatusWithOutlierDetection(),
					},
					&discoveryv1alpha2.MeshService{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "meshservice1",
							Namespace:   "namespace1",
							ClusterName: "cluster3",
						},
						Spec: discoveryv1alpha2.MeshServiceSpec{
							Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
								KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
									Ref: &corev1.ClusterObjectRef{
										Name:        "service1",
										Namespace:   "namespace1",
										ClusterName: "cluster3",
									},
								},
							},
							Mesh: &corev1.ObjectRef{
								Name:      "mesh3",
								Namespace: "namespace1",
							},
						},
						Status: meshServiceStatusWithOutlierDetection(),
					},
					&discoveryv1alpha2.MeshService{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "meshservice1",
							Namespace:   "namespace1",
							ClusterName: "cluster4",
						},
						Spec: discoveryv1alpha2.MeshServiceSpec{
							Type: &discoveryv1alpha2.MeshServiceSpec_KubeService_{
								KubeService: &discoveryv1alpha2.MeshServiceSpec_KubeService{
									Ref: &corev1.ClusterObjectRef{
										Name:        "service1",
										Namespace:   "namespace1",
										ClusterName: "cluster4",
									},
								},
							},
							Mesh: &corev1.ObjectRef{
								Name:      "mesh3",
								Namespace: "namespace1",
							},
						},
						Status: discoveryv1alpha2.MeshServiceStatus{
							AppliedTrafficPolicies: []*discoveryv1alpha2.MeshServiceStatus_AppliedTrafficPolicy{
								{
									// Missing OutlierDetection
									Spec: &networkingv1alpha2.TrafficPolicySpec{},
								},
							},
						},
					},
				),
				KubeClusters: multiclusterv1alpha1sets.NewKubernetesClusterSet(),
				Meshes: discoveryv1alpha2sets.NewMeshSet(
					&discoveryv1alpha2.Mesh{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mesh1",
							Namespace: "namespace1",
						},
						Spec: discoveryv1alpha2.MeshSpec{
							MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
								Istio: &discoveryv1alpha2.MeshSpec_Istio{
									Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
										Cluster: "cluster1",
									},
								},
							},
						},
						Status: discoveryv1alpha2.MeshStatus{
							AppliedVirtualMeshes: []*discoveryv1alpha2.MeshStatus_AppliedVirtualMesh{
								{
									Ref: &corev1.ObjectRef{
										Name:      "vm1",
										Namespace: "namespace1",
									},
								},
							},
						},
					},
					&discoveryv1alpha2.Mesh{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mesh2",
							Namespace: "namespace1",
						},
						Spec: discoveryv1alpha2.MeshSpec{
							MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
								Istio: &discoveryv1alpha2.MeshSpec_Istio{
									Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
										Cluster: "cluster2",
									},
								},
							},
						},
						Status: discoveryv1alpha2.MeshStatus{
							AppliedVirtualMeshes: []*discoveryv1alpha2.MeshStatus_AppliedVirtualMesh{
								{
									Ref: &corev1.ObjectRef{
										Name:      "vm2",
										Namespace: "namespace1",
									},
								},
							},
						},
					},
					&discoveryv1alpha2.Mesh{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mesh3",
							Namespace: "namespace1",
						},
						Spec: discoveryv1alpha2.MeshSpec{
							MeshType: &discoveryv1alpha2.MeshSpec_Istio_{
								Istio: &discoveryv1alpha2.MeshSpec_Istio{
									Installation: &discoveryv1alpha2.MeshSpec_MeshInstallation{
										Cluster: "cluster3",
									},
								},
							},
						},
					},
				),
				VirtualMeshes: networkingv1alpha2sets.NewVirtualMeshSet(
					&networkingv1alpha2.VirtualMesh{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "vm1",
							Namespace: "namespace1",
						},
						Spec: networkingv1alpha2.VirtualMeshSpec{
							Meshes: []*corev1.ObjectRef{
								{
									Name:      "mesh1",
									Namespace: "namespace1",
								},
							},
						},
					},
					&networkingv1alpha2.VirtualMesh{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "vm2",
							Namespace: "namespace1",
						},
						Spec: networkingv1alpha2.VirtualMeshSpec{
							Meshes: []*corev1.ObjectRef{
								{
									Name:      "mesh2",
									Namespace: "namespace1",
								},
							},
						},
					},
				),
			}, &networkingv1alpha2.FailoverServiceSpec{
				Hostname: "invalidDNS@Q#$@%",
				Meshes: []*corev1.ObjectRef{
					{
						Name:      "mesh1",
						Namespace: "namespace1",
					},
				},
				FailoverServices: []*corev1.ClusterObjectRef{
					{
						Name:        "service1",
						Namespace:   "namespace1",
						ClusterName: "cluster1",
					},
					{
						Name:        "service1",
						Namespace:   "namespace1",
						ClusterName: "cluster2",
					},
					{
						Name:        "service1",
						Namespace:   "namespace1",
						ClusterName: "cluster3",
					},
					{
						Name:        "service1",
						Namespace:   "namespace1",
						ClusterName: "cluster4",
					},
				},
			}
	}

	It("should return no errors for valid FailoverService", func() {
		inputSnapshot, failoverServiceSpec := validInputs()
		err := validator.Validate(inputSnapshot, failoverServiceSpec)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return no errors for FailoverService composed of services belonging to a single common mesh, with no VirtualMesh", func() {
		inputSnapshot, failoverServiceSpec := validInputSnapshotSingleMesh()
		err := validator.Validate(inputSnapshot, failoverServiceSpec)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return errors for invalid FailoverService", func() {
		inputSnapshot, failoverServiceSpec := invalidInputSnapshot()
		err := validator.Validate(inputSnapshot, failoverServiceSpec)
		// Missing port
		Expect(err.Error()).To(ContainSubstring(validation.MissingPort.Error()))
		// Invalid DNS hostname
		Expect(err.Error()).To(ContainSubstring("a DNS-1123 subdomain must consist of lower case alphanumeric characters"))
		// Service without OutlierDetection
		Expect(err.Error()).To(ContainSubstring(validation.MissingOutlierDetection(inputSnapshot.MeshServices.List()[3]).Error()))
		// Mesh without parent VirtualMesh
		Expect(err.Error()).To(ContainSubstring(
			validation.MeshWithoutParentVM(inputSnapshot.Meshes.List()[2]).Error()))
		// Multiple parent VirtualMeshes
		Expect(err.Error()).To(ContainSubstring(validation.MultipleParentVirtualMeshes(inputSnapshot.VirtualMeshes.List()).Error()))
	})
})
