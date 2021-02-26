package apply_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/gloo-mesh/pkg/mesh-networking/apply"
)

var _ = Describe("Applier", func() {
	Context("applied traffic policies", func() {
		var (
			destination = &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
				Spec: discoveryv1.DestinationSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "svc-name",
								Namespace:   "svc-namespace",
								ClusterName: "svc-cluster",
							},
						},
					},
				},
			}
			workload = &discoveryv1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld1",
					Namespace: "ns",
				},
				Spec: discoveryv1.WorkloadSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
				},
			}
			mesh = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "ns",
				},
			}
			trafficPolicy1 = &v1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
				Spec: v1.TrafficPolicySpec{
					Policy: &v1.TrafficPolicySpec_Policy{
						// fill an arbitrary part of the spec
						Mirror: &v1.TrafficPolicySpec_Policy_Mirror{},
					},
				},
			}
			trafficPolicy2 = &v1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp2",
					Namespace: "ns",
				},
				Spec: v1.TrafficPolicySpec{
					Policy: &v1.TrafficPolicySpec_Policy{
						// fill an arbitrary part of the spec
						FaultInjection: &v1.TrafficPolicySpec_Policy_FaultInjection{},
					},
				},
			}

			snap = input.NewInputLocalSnapshotManualBuilder("").
				AddDestinations(discoveryv1.DestinationSlice{destination}).
				AddTrafficPolicies(v1.TrafficPolicySlice{trafficPolicy1, trafficPolicy2}).
				AddWorkloads(discoveryv1.WorkloadSlice{workload}).
				AddMeshes(discoveryv1.MeshSlice{mesh}).
				Build()
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil)
		})
		It("updates status on input traffic policies", func() {
			Expect(trafficPolicy1.Status.Destinations).To(HaveKey(sets.Key(destination)))
			Expect(trafficPolicy1.Status.Destinations[sets.Key(destination)]).To(Equal(&v1.ApprovalStatus{
				AcceptanceOrder: 0,
				State:           commonv1.ApprovalState_ACCEPTED,
			}))
			Expect(trafficPolicy1.Status.Workloads).To(HaveLen(1))
			Expect(trafficPolicy1.Status.Workloads[0]).To(Equal(sets.Key(workload)))
			Expect(trafficPolicy2.Status.Destinations).To(HaveKey(sets.Key(destination)))
			Expect(trafficPolicy2.Status.Destinations[sets.Key(destination)]).To(Equal(&v1.ApprovalStatus{
				AcceptanceOrder: 1,
				State:           commonv1.ApprovalState_ACCEPTED,
			}))
			Expect(trafficPolicy2.Status.Workloads).To(HaveLen(1))
			Expect(trafficPolicy2.Status.Workloads[0]).To(Equal(sets.Key(workload)))

		})
		It("updates status on input Destination policies", func() {
			Expect(destination.Status.AppliedTrafficPolicies).To(HaveLen(2))
			Expect(destination.Status.AppliedTrafficPolicies[0].Ref).To(Equal(ezkube.MakeObjectRef(trafficPolicy1)))
			Expect(destination.Status.AppliedTrafficPolicies[0].Spec).To(Equal(&trafficPolicy1.Spec))
			Expect(destination.Status.AppliedTrafficPolicies[1].Ref).To(Equal(ezkube.MakeObjectRef(trafficPolicy2)))
			Expect(destination.Status.AppliedTrafficPolicies[1].Spec).To(Equal(&trafficPolicy2.Spec))
			Expect(destination.Status.LocalFqdn).To(Equal("svc-name.svc-namespace.svc.cluster.local"))
		})
	})
	Context("invalid traffic policies", func() {
		var (
			destination = &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
			}
			trafficPolicy = &v1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
			}

			snap = input.NewInputLocalSnapshotManualBuilder("").
				AddDestinations(discoveryv1.DestinationSlice{destination}).
				AddTrafficPolicies(v1.TrafficPolicySlice{trafficPolicy}).
				Build()
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// report = reject
				reporter.ReportTrafficPolicyToDestination(destination, trafficPolicy, errors.New("did an oopsie"))
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil)
		})
		It("updates status on input traffic policies", func() {
			Expect(trafficPolicy.Status.Destinations).To(HaveKey(sets.Key(destination)))
			Expect(trafficPolicy.Status.Destinations[sets.Key(destination)]).To(Equal(&v1.ApprovalStatus{
				AcceptanceOrder: 0,
				State:           commonv1.ApprovalState_INVALID,
				Errors:          []string{"did an oopsie"},
			}))
		})
		It("does not add the policy to the Destination status", func() {
			Expect(destination.Status.AppliedTrafficPolicies).To(HaveLen(0))
		})
	})

	Context("setting workloads status", func() {
		var (
			destination = &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
				Spec: discoveryv1.DestinationSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
				},
			}
			workload1 = &discoveryv1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld1",
					Namespace: "ns",
				},
				Spec: discoveryv1.WorkloadSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
				},
			}
			workload2 = &discoveryv1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld2",
					Namespace: "ns",
				},
				Spec: discoveryv1.WorkloadSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh2",
						Namespace: "ns",
					},
				},
			}
			mesh1 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "ns",
				},
			}
			mesh2 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh2",
					Namespace: "ns",
				},
			}
			virtualMesh = &v1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vmesh1",
					Namespace: "ns",
				},
				Spec: v1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						{Name: "mesh1", Namespace: "ns"},
						{Name: "mesh2", Namespace: "ns"},
					},
				},
			}
			trafficPolicy = &v1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
			}
			accessPolicy = &v1.AccessPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ap1",
					Namespace: "ns",
				},
			}
		)

		It("sets policy workloads using mesh", func() {
			snap := input.NewInputLocalSnapshotManualBuilder("").
				AddTrafficPolicies(v1.TrafficPolicySlice{trafficPolicy}).
				AddAccessPolicies(v1.AccessPolicySlice{accessPolicy}).
				AddDestinations(discoveryv1.DestinationSlice{destination}).
				AddWorkloads(discoveryv1.WorkloadSlice{workload1, workload2}).
				AddMeshes(discoveryv1.MeshSlice{mesh1, mesh2}).
				Build()
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil)

			// destination and workload1 are both in mesh1
			Expect(trafficPolicy.Status.Workloads).To(HaveLen(1))
			Expect(trafficPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
			Expect(accessPolicy.Status.Workloads).To(HaveLen(1))
			Expect(accessPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
		})
		It("sets policy workloads using VirtualMesh", func() {
			snap := input.NewInputLocalSnapshotManualBuilder("").
				AddTrafficPolicies(v1.TrafficPolicySlice{trafficPolicy}).
				AddAccessPolicies(v1.AccessPolicySlice{accessPolicy}).
				AddDestinations(discoveryv1.DestinationSlice{destination}).
				AddWorkloads(discoveryv1.WorkloadSlice{workload1, workload2}).
				AddMeshes(discoveryv1.MeshSlice{mesh1, mesh2}).
				AddVirtualMeshes(v1.VirtualMeshSlice{virtualMesh}).
				Build()
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil)

			// destination is in mesh1, workload1 is in mesh1, and workload2 is in mesh2.
			// since mesh1 and mesh2 are in the same VirtualMesh, both workloads are returned
			Expect(trafficPolicy.Status.Workloads).To(HaveLen(2))
			Expect(trafficPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
			Expect(trafficPolicy.Status.Workloads[1]).To(Equal(sets.Key(workload2)))
			Expect(accessPolicy.Status.Workloads).To(HaveLen(2))
			Expect(accessPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
			Expect(accessPolicy.Status.Workloads[1]).To(Equal(sets.Key(workload2)))
		})
		It("sets no policy workloads when there is no matching mesh", func() {
			workload1.Spec.Mesh.Name = "mesh2"
			snap := input.NewInputLocalSnapshotManualBuilder("").
				AddTrafficPolicies(v1.TrafficPolicySlice{trafficPolicy}).
				AddAccessPolicies(v1.AccessPolicySlice{accessPolicy}).
				AddDestinations(discoveryv1.DestinationSlice{destination}).
				AddWorkloads(discoveryv1.WorkloadSlice{workload1, workload2}).
				AddMeshes(discoveryv1.MeshSlice{mesh1, mesh2}).
				Build()
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil)

			// destination is in mesh1, but both workloads are in mesh2
			Expect(trafficPolicy.Status.Workloads).To(BeNil())
			Expect(accessPolicy.Status.Workloads).To(BeNil())
		})
	})
})

// NOTE(ilackarms): we implement a test translator here instead of using a mock because
// we need to call methods on the reporter which is passed as an argument to the translator
type testIstioTranslator struct {
	callReporter func(reporter reporting.Reporter)
}

func (t testIstioTranslator) Translate(
	ctx context.Context,
	in input.LocalSnapshot,
	existingIstioResources input.RemoteSnapshot,
	reporter reporting.Reporter,
) (*translation.Outputs, error) {
	t.callReporter(reporter)
	return &translation.Outputs{}, nil
}
