package configtarget_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	commonv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1alpha2"
	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2/sets"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/apply/configtarget"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
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
		meshes := discoveryv1alpha2sets.NewMeshSet(
			&discoveryv1alpha2.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			})
		destinations := discoveryv1alpha2sets.NewDestinationSet(
			&discoveryv1alpha2.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "destination",
					Namespace: "namespace",
				},
				Spec: discoveryv1alpha2.DestinationSpec{
					Type: &discoveryv1alpha2.DestinationSpec_KubeService_{
						KubeService: &discoveryv1alpha2.DestinationSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "foo",
								Namespace:   "bar",
								ClusterName: "cluster",
							},
						},
					},
				},
			})

		validator = configtarget.NewConfigTargetValidator(meshes, destinations)

		accessPolicies := v1alpha2.AccessPolicySlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid",
					Namespace: namespace,
				},
				Spec: v1alpha2.AccessPolicySpec{
					DestinationSelector: []*commonv1alpha2.DestinationSelector{
						{
							KubeServiceRefs: &commonv1alpha2.DestinationSelector_KubeServiceRefs{
								Services: []*v1.ClusterObjectRef{
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
				Status: v1alpha2.AccessPolicyStatus{
					State: v1alpha2.ApprovalState_ACCEPTED,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: namespace,
				},
				Spec: v1alpha2.AccessPolicySpec{
					DestinationSelector: []*commonv1alpha2.DestinationSelector{
						{
							KubeServiceRefs: &commonv1alpha2.DestinationSelector_KubeServiceRefs{
								Services: []*v1.ClusterObjectRef{
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
				Status: v1alpha2.AccessPolicyStatus{
					State: v1alpha2.ApprovalState_ACCEPTED,
				},
			},
		}

		trafficPolicies := v1alpha2.TrafficPolicySlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid",
					Namespace: namespace,
				},
				Spec: v1alpha2.TrafficPolicySpec{
					DestinationSelector: []*commonv1alpha2.DestinationSelector{
						{
							KubeServiceRefs: &commonv1alpha2.DestinationSelector_KubeServiceRefs{
								Services: []*v1.ClusterObjectRef{
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
				Status: v1alpha2.TrafficPolicyStatus{
					State: v1alpha2.ApprovalState_ACCEPTED,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: namespace,
				},
				Spec: v1alpha2.TrafficPolicySpec{
					DestinationSelector: []*commonv1alpha2.DestinationSelector{
						{
							KubeServiceRefs: &commonv1alpha2.DestinationSelector_KubeServiceRefs{
								Services: []*v1.ClusterObjectRef{
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
				Status: v1alpha2.TrafficPolicyStatus{
					State: v1alpha2.ApprovalState_ACCEPTED,
				},
			},
		}

		virtualMeshes := v1alpha2.VirtualMeshSlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "valid",
					Namespace: namespace,
				},
				Spec: v1alpha2.VirtualMeshSpec{
					Meshes: []*v1.ObjectRef{
						{
							Name:      "foo",
							Namespace: "bar",
						},
					},
				},
				Status: v1alpha2.VirtualMeshStatus{
					State: v1alpha2.ApprovalState_ACCEPTED,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid",
					Namespace: namespace,
				},
				Spec: v1alpha2.VirtualMeshSpec{
					Meshes: []*v1.ObjectRef{
						{
							Name:      "nonexistent",
							Namespace: "nonexistent",
						},
					},
				},
				Status: v1alpha2.VirtualMeshStatus{
					State: v1alpha2.ApprovalState_ACCEPTED,
				},
			},
		}

		validator.ValidateAccessPolicies(accessPolicies)
		validator.ValidateTrafficPolicies(trafficPolicies)
		validator.ValidateVirtualMeshes(virtualMeshes)

		Expect(accessPolicies[0].Status.State).To(Equal(v1alpha2.ApprovalState_ACCEPTED))
		Expect(trafficPolicies[0].Status.State).To(Equal(v1alpha2.ApprovalState_ACCEPTED))
		Expect(virtualMeshes[0].Status.State).To(Equal(v1alpha2.ApprovalState_ACCEPTED))

		Expect(accessPolicies[1].Status.State).To(Equal(v1alpha2.ApprovalState_INVALID))
		Expect(trafficPolicies[1].Status.State).To(Equal(v1alpha2.ApprovalState_INVALID))
		Expect(virtualMeshes[1].Status.State).To(Equal(v1alpha2.ApprovalState_INVALID))
	})

	It("should validate one VirtualMesh per mesh", func() {
		meshes := discoveryv1alpha2sets.NewMeshSet(
			&discoveryv1alpha2.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "namespace1",
				},
			},
			&discoveryv1alpha2.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh2",
					Namespace: "namespace1",
				},
			},
			&discoveryv1alpha2.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh3",
					Namespace: "namespace1",
				},
			},
		)

		validator = configtarget.NewConfigTargetValidator(meshes, nil)

		vm1 := &v1alpha2.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm1",
				Namespace: "namespace1",
			},
			Spec: v1alpha2.VirtualMeshSpec{
				Meshes: []*v1.ObjectRef{
					{
						Name:      "mesh1",
						Namespace: "namespace1",
					},
				},
			},
			Status: v1alpha2.VirtualMeshStatus{
				State: v1alpha2.ApprovalState_ACCEPTED,
			},
		}
		vm2 := &v1alpha2.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm2",
				Namespace: "namespace1",
			},
			Spec: v1alpha2.VirtualMeshSpec{
				Meshes: []*v1.ObjectRef{
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
			Status: v1alpha2.VirtualMeshStatus{
				State: v1alpha2.ApprovalState_ACCEPTED,
			},
		}
		vm3 := &v1alpha2.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm3",
				Namespace: "namespace1",
			},
			Spec: v1alpha2.VirtualMeshSpec{
				Meshes: []*v1.ObjectRef{
					{
						Name:      "mesh2",
						Namespace: "namespace1",
					},
				},
			},
			Status: v1alpha2.VirtualMeshStatus{
				State: v1alpha2.ApprovalState_ACCEPTED,
			},
		}
		vm4 := &v1alpha2.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm4",
				Namespace: "namespace1",
			},
			Spec: v1alpha2.VirtualMeshSpec{
				Meshes: []*v1.ObjectRef{
					{
						Name:      "mesh2",
						Namespace: "namespace1",
					},
				},
			},
			Status: v1alpha2.VirtualMeshStatus{
				State: v1alpha2.ApprovalState_ACCEPTED,
			},
		}
		vm5 := &v1alpha2.VirtualMesh{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vm5",
				Namespace: "namespace1",
			},
			Spec: v1alpha2.VirtualMeshSpec{
				Meshes: []*v1.ObjectRef{
					{
						Name:      "mesh3",
						Namespace: "namespace1",
					},
				},
			},
			Status: v1alpha2.VirtualMeshStatus{
				State: v1alpha2.ApprovalState_ACCEPTED,
			},
		}

		validator.ValidateVirtualMeshes(v1alpha2.VirtualMeshSlice{vm5, vm4, vm3, vm2, vm1})

		Expect(vm2.Status.State).To(Equal(v1alpha2.ApprovalState_INVALID))
		Expect(vm3.Status.State).To(Equal(v1alpha2.ApprovalState_ACCEPTED))
		Expect(vm5.Status.State).To(Equal(v1alpha2.ApprovalState_ACCEPTED))
	})
})
