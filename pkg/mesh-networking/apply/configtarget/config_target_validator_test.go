package configtarget_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	discoveryv1sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1/sets"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/apply/configtarget"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ConfigTargetValidator", func() {
	var (
		ctrl      *gomock.Controller
		namespace = "policy-namespace"
		validator configtarget.ConfigTargetValidator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should invalidate any policies that reference non-existent discovery entities", func() {
		meshes := discoveryv1sets.NewMeshSet(
			&discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			})
		destinations := discoveryv1sets.NewDestinationSet(
			&discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "destination",
					Namespace: "namespace",
				},
				Spec: discoveryv1.DestinationSpec{
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "foo",
								Namespace:   "bar",
								ClusterName: "cluster",
							},
							WorkloadSelectorLabels: map[string]string{"istio": "ingressgateway"},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
								{
									Port: 1234,
									Name: defaults.IstioGatewayTlsPortName,
								},
							},
						},
					},
					Mesh: &skv2corev1.ObjectRef{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			})

		validator = configtarget.NewConfigTargetValidator(meshes, destinations)

		accessPolicies := v1.AccessPolicySlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid",
					Namespace: namespace,
				},
				Spec: v1.AccessPolicySpec{
					DestinationSelector: []*commonv1.DestinationSelector{
						{
							KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
								Services: []*skv2corev1.ClusterObjectRef{
									{
										Name:        "foo",
										Namespace:   "bar",
										ClusterName: "cluster",
									},
								},
							},
						},
					},
				},
				Status: v1.AccessPolicyStatus{
					State: commonv1.ApprovalState_ACCEPTED,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: namespace,
				},
				Spec: v1.AccessPolicySpec{
					DestinationSelector: []*commonv1.DestinationSelector{
						{
							KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
								Services: []*skv2corev1.ClusterObjectRef{
									{
										Name:        "nonexistent",
										Namespace:   "nonexistent",
										ClusterName: "nonexistent",
									},
								},
							},
						},
					},
				},
				Status: v1.AccessPolicyStatus{
					State: commonv1.ApprovalState_ACCEPTED,
				},
			},
		}

		trafficPolicies := v1.TrafficPolicySlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid",
					Namespace: namespace,
				},
				Spec: v1.TrafficPolicySpec{
					DestinationSelector: []*commonv1.DestinationSelector{
						{
							KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
								Services: []*skv2corev1.ClusterObjectRef{
									{
										Name:        "foo",
										Namespace:   "bar",
										ClusterName: "cluster",
									},
								},
							},
						},
					},
				},
				Status: v1.TrafficPolicyStatus{
					State: commonv1.ApprovalState_ACCEPTED,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: namespace,
				},
				Spec: v1.TrafficPolicySpec{
					DestinationSelector: []*commonv1.DestinationSelector{
						{
							KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
								Services: []*skv2corev1.ClusterObjectRef{
									{
										Name:        "nonexistent",
										Namespace:   "nonexistent",
										ClusterName: "nonexistent",
									},
								},
							},
						},
					},
				},
				Status: v1.TrafficPolicyStatus{
					State: commonv1.ApprovalState_ACCEPTED,
				},
			},
		}

		virtualMeshes := v1.VirtualMeshSlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid",
					Namespace: namespace,
				},
				Spec: v1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						{
							Name:      "foo",
							Namespace: "bar",
						},
					},
				},
				Status: v1.VirtualMeshStatus{
					State: commonv1.ApprovalState_ACCEPTED,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: namespace,
				},
				Spec: v1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						{
							Name:      "nonexistent",
							Namespace: "nonexistent",
						},
					},
				},
				Status: v1.VirtualMeshStatus{
					State: commonv1.ApprovalState_ACCEPTED,
				},
			},
		}

		validator.ValidateAccessPolicies(accessPolicies)
		validator.ValidateTrafficPolicies(trafficPolicies)
		validator.ValidateVirtualMeshes(virtualMeshes)

		Expect(accessPolicies[0].Status.State).To(Equal(commonv1.ApprovalState_ACCEPTED))
		Expect(trafficPolicies[0].Status.State).To(Equal(commonv1.ApprovalState_ACCEPTED))
		Expect(virtualMeshes[0].Status.State).To(Equal(commonv1.ApprovalState_ACCEPTED))

		Expect(accessPolicies[1].Status.State).To(Equal(commonv1.ApprovalState_INVALID))
		Expect(trafficPolicies[1].Status.State).To(Equal(commonv1.ApprovalState_INVALID))
		Expect(virtualMeshes[1].Status.State).To(Equal(commonv1.ApprovalState_INVALID))
	})

	It("should validate one VirtualMesh per mesh", func() {
		meshes := discoveryv1sets.NewMeshSet(
			&discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "namespace1",
				},
			},
			&discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh2",
					Namespace: "namespace1",
				},
			},
			&discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh3",
					Namespace: "namespace1",
				},
			},
		)

		destination1 := &discoveryv1.Destination{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "destination",
				Namespace: "namespace",
			},
			Spec: discoveryv1.DestinationSpec{
				Type: &discoveryv1.DestinationSpec_KubeService_{
					KubeService: &discoveryv1.DestinationSpec_KubeService{
						Ref: &skv2corev1.ClusterObjectRef{
							Name:        "foo",
							Namespace:   "bar",
							ClusterName: "cluster",
						},
						WorkloadSelectorLabels: defaults.DefaultGatewayWorkloadLabels,
						Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
							{
								Port: 1234,
								Name: defaults.IstioGatewayTlsPortName,
							},
						},
					},
				},
				Mesh: &skv2corev1.ObjectRef{
					Name:      "mesh1",
					Namespace: "namespace1",
				},
			},
		}
		destination2 := destination1.DeepCopy()
		destination2.Name = "destination2"
		destination2.Spec.Mesh = &skv2corev1.ObjectRef{
			Name:      "mesh2",
			Namespace: "namespace1",
		}
		destination3 := destination1.DeepCopy()
		destination3.Name = "destination3"
		destination3.Spec.Mesh = &skv2corev1.ObjectRef{
			Name:      "mesh3",
			Namespace: "namespace1",
		}
		destinations := discoveryv1sets.NewDestinationSet(
			destination1, destination2, destination3)

		validator = configtarget.NewConfigTargetValidator(meshes, destinations)

		vm1 := &v1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm1",
				Namespace: "namespace1",
			},
			Spec: v1.VirtualMeshSpec{
				Meshes: []*skv2corev1.ObjectRef{
					{
						Name:      "mesh1",
						Namespace: "namespace1",
					},
				},
			},
			Status: v1.VirtualMeshStatus{
				State: commonv1.ApprovalState_ACCEPTED,
			},
		}
		vm2 := &v1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm2",
				Namespace: "namespace1",
			},
			Spec: v1.VirtualMeshSpec{
				Meshes: []*skv2corev1.ObjectRef{
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
			Status: v1.VirtualMeshStatus{
				State: commonv1.ApprovalState_ACCEPTED,
			},
		}
		vm3 := &v1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm3",
				Namespace: "namespace1",
			},
			Spec: v1.VirtualMeshSpec{
				Meshes: []*skv2corev1.ObjectRef{
					{
						Name:      "mesh2",
						Namespace: "namespace1",
					},
				},
			},
			Status: v1.VirtualMeshStatus{
				State: commonv1.ApprovalState_ACCEPTED,
			},
		}
		vm4 := &v1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm4",
				Namespace: "namespace1",
			},
			Spec: v1.VirtualMeshSpec{
				Meshes: []*skv2corev1.ObjectRef{
					{
						Name:      "mesh2",
						Namespace: "namespace1",
					},
				},
			},
			Status: v1.VirtualMeshStatus{
				State: commonv1.ApprovalState_ACCEPTED,
			},
		}
		vm5 := &v1.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm5",
				Namespace: "namespace1",
			},
			Spec: v1.VirtualMeshSpec{
				Meshes: []*skv2corev1.ObjectRef{
					{
						Name:      "mesh3",
						Namespace: "namespace1",
					},
				},
			},
			Status: v1.VirtualMeshStatus{
				State: commonv1.ApprovalState_ACCEPTED,
			},
		}

		validator.ValidateVirtualMeshes(v1.VirtualMeshSlice{vm5, vm4, vm3, vm2, vm1})

		Expect(vm2.Status.State).To(Equal(commonv1.ApprovalState_INVALID))
		Expect(vm3.Status.State).To(Equal(commonv1.ApprovalState_ACCEPTED))
		Expect(vm5.Status.State).To(Equal(commonv1.ApprovalState_ACCEPTED))
	})

	It("should invalidate any policies that don't specify all fields in a reference-based selector", func() {
		validator = configtarget.NewConfigTargetValidator(nil, nil)

		trafficPolicies := v1.TrafficPolicySlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid",
					Namespace: namespace,
				},
				Spec: v1.TrafficPolicySpec{
					DestinationSelector: []*commonv1.DestinationSelector{
						{
							KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
								Services: []*skv2corev1.ClusterObjectRef{
									{
										// missing name
										// missing namespace
										// missing clustername
									},
								},
							},
						},
					},
				},
				Status: v1.TrafficPolicyStatus{
					State: commonv1.ApprovalState_ACCEPTED,
				},
			},
		}

		virtualMeshes := v1.VirtualMeshSlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid",
					Namespace: namespace,
				},
				Spec: v1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						{
							// missing name
							// missing namespace
						},
					},
				},
				Status: v1.VirtualMeshStatus{
					State: commonv1.ApprovalState_ACCEPTED,
				},
			},
		}

		validator.ValidateTrafficPolicies(trafficPolicies)
		validator.ValidateVirtualMeshes(virtualMeshes)

		Expect(trafficPolicies[0].Status.State).To(Equal(commonv1.ApprovalState_INVALID))
		Expect(trafficPolicies[0].Status.Errors).To(Equal([]string{"malformed kubeServiceRef: 3 errors occurred:\n\t* 'name' must be specified'\n\t* 'namespace' must be specified'\n\t* 'clusterName' must be specified'\n\n"}))
		Expect(virtualMeshes[0].Status.State).To(Equal(commonv1.ApprovalState_INVALID))
		Expect(virtualMeshes[0].Status.Errors).To(Equal([]string{"malformed meshRef: 2 errors occurred:\n\t* 'name' must be specified'\n\t* 'namespace' must be specified'\n\n"}))
	})
})
