package apply_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/apply"
)

var _ = Describe("Applier", func() {
	Context("applied traffic policies", func() {
		var (
			trafficTarget = &discoveryv1alpha2.TrafficTarget{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Mesh: &v1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
					Type: &discoveryv1alpha2.TrafficTargetSpec_KubeService_{
						KubeService: &discoveryv1alpha2.TrafficTargetSpec_KubeService{
							Ref: &v1.ClusterObjectRef{
								Name:        "svc-name",
								Namespace:   "svc-namespace",
								ClusterName: "svc-cluster",
							},
						},
					},
				},
			}
			workload = &discoveryv1alpha2.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld1",
					Namespace: "ns",
				},
				Spec: discoveryv1alpha2.WorkloadSpec{
					Mesh: &v1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
				},
			}
			mesh = &discoveryv1alpha2.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "ns",
				},
			}
			trafficPolicy1 = &v1alpha2.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
				Spec: v1alpha2.TrafficPolicySpec{
					// fill an arbitrary part of the spec
					Mirror: &v1alpha2.TrafficPolicySpec_Mirror{},
				},
			}
			trafficPolicy2 = &v1alpha2.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp2",
					Namespace: "ns",
				},
				Spec: v1alpha2.TrafficPolicySpec{
					// fill an arbitrary part of the spec
					FaultInjection: &v1alpha2.TrafficPolicySpec_FaultInjection{},
				},
			}

			snap = input.NewInputSnapshotManualBuilder("").
				AddTrafficTargets(discoveryv1alpha2.TrafficTargetSlice{trafficTarget}).
				AddTrafficPolicies(v1alpha2.TrafficPolicySlice{trafficPolicy1, trafficPolicy2}).
				AddWorkloads(discoveryv1alpha2.WorkloadSlice{workload}).
				AddMeshes(discoveryv1alpha2.MeshSlice{mesh}).
				Build()
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap)
		})
		It("updates status on input traffic policies", func() {
			Expect(trafficPolicy1.Status.TrafficTargets).To(HaveKey(sets.Key(trafficTarget)))
			Expect(trafficPolicy1.Status.TrafficTargets[sets.Key(trafficTarget)]).To(Equal(&v1alpha2.ApprovalStatus{
				AcceptanceOrder: 0,
				State:           v1alpha2.ApprovalState_ACCEPTED,
			}))
			Expect(trafficPolicy1.Status.Workloads).To(HaveLen(1))
			Expect(trafficPolicy1.Status.Workloads[0]).To(Equal(sets.Key(workload)))
			Expect(trafficPolicy2.Status.TrafficTargets).To(HaveKey(sets.Key(trafficTarget)))
			Expect(trafficPolicy2.Status.TrafficTargets[sets.Key(trafficTarget)]).To(Equal(&v1alpha2.ApprovalStatus{
				AcceptanceOrder: 1,
				State:           v1alpha2.ApprovalState_ACCEPTED,
			}))
			Expect(trafficPolicy2.Status.Workloads).To(HaveLen(1))
			Expect(trafficPolicy2.Status.Workloads[0]).To(Equal(sets.Key(workload)))

		})
		It("updates status on input traffic target policies", func() {
			Expect(trafficTarget.Status.AppliedTrafficPolicies).To(HaveLen(2))
			Expect(trafficTarget.Status.AppliedTrafficPolicies[0].Ref).To(Equal(ezkube.MakeObjectRef(trafficPolicy1)))
			Expect(trafficTarget.Status.AppliedTrafficPolicies[0].Spec).To(Equal(&trafficPolicy1.Spec))
			Expect(trafficTarget.Status.AppliedTrafficPolicies[1].Ref).To(Equal(ezkube.MakeObjectRef(trafficPolicy2)))
			Expect(trafficTarget.Status.AppliedTrafficPolicies[1].Spec).To(Equal(&trafficPolicy2.Spec))
			Expect(trafficTarget.Status.LocalFqdn).To(Equal("svc-name.svc-namespace.svc.cluster.local"))
			Expect(trafficTarget.Status.RemoteFqdn).To(Equal("svc-name.svc-namespace.svc.svc-cluster.global"))
		})
	})
	Context("invalid traffic policies", func() {
		var (
			trafficTarget = &discoveryv1alpha2.TrafficTarget{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
			}
			trafficPolicy = &v1alpha2.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
			}

			snap = input.NewInputSnapshotManualBuilder("").
				AddTrafficTargets(discoveryv1alpha2.TrafficTargetSlice{trafficTarget}).
				AddTrafficPolicies(v1alpha2.TrafficPolicySlice{trafficPolicy}).
				Build()
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// report = reject
				reporter.ReportTrafficPolicyToTrafficTarget(trafficTarget, trafficPolicy, errors.New("did an oopsie"))
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap)
		})
		It("updates status on input traffic policies", func() {
			Expect(trafficPolicy.Status.TrafficTargets).To(HaveKey(sets.Key(trafficTarget)))
			Expect(trafficPolicy.Status.TrafficTargets[sets.Key(trafficTarget)]).To(Equal(&v1alpha2.ApprovalStatus{
				AcceptanceOrder: 0,
				State:           v1alpha2.ApprovalState_INVALID,
				Errors:          []string{"did an oopsie"},
			}))
		})
		It("does not add the policy to the traffic target status", func() {
			Expect(trafficTarget.Status.AppliedTrafficPolicies).To(HaveLen(0))
		})
	})

	Context("setting workloads status", func() {
		var (
			trafficTarget = &discoveryv1alpha2.TrafficTarget{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
				Spec: discoveryv1alpha2.TrafficTargetSpec{
					Mesh: &v1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
				},
			}
			workload1 = &discoveryv1alpha2.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld1",
					Namespace: "ns",
				},
				Spec: discoveryv1alpha2.WorkloadSpec{
					Mesh: &v1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
				},
			}
			workload2 = &discoveryv1alpha2.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld2",
					Namespace: "ns",
				},
				Spec: discoveryv1alpha2.WorkloadSpec{
					Mesh: &v1.ObjectRef{
						Name:      "mesh2",
						Namespace: "ns",
					},
				},
			}
			mesh1 = &discoveryv1alpha2.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "ns",
				},
			}
			mesh2 = &discoveryv1alpha2.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh2",
					Namespace: "ns",
				},
			}
			virtualMesh = &v1alpha2.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vmesh1",
					Namespace: "ns",
				},
				Spec: v1alpha2.VirtualMeshSpec{
					Meshes: []*v1.ObjectRef{
						{Name: "mesh1", Namespace: "ns"},
						{Name: "mesh2", Namespace: "ns"},
					},
				},
			}
			trafficPolicy = &v1alpha2.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
			}
			accessPolicy = &v1alpha2.AccessPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ap1",
					Namespace: "ns",
				},
			}
		)

		It("sets policy workloads using mesh", func() {
			snap := input.NewInputSnapshotManualBuilder("").
				AddTrafficPolicies(v1alpha2.TrafficPolicySlice{trafficPolicy}).
				AddAccessPolicies(v1alpha2.AccessPolicySlice{accessPolicy}).
				AddTrafficTargets(discoveryv1alpha2.TrafficTargetSlice{trafficTarget}).
				AddWorkloads(discoveryv1alpha2.WorkloadSlice{workload1, workload2}).
				AddMeshes(discoveryv1alpha2.MeshSlice{mesh1, mesh2}).
				Build()
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap)

			// trafficTarget and workload1 are both in mesh1
			Expect(trafficPolicy.Status.Workloads).To(HaveLen(1))
			Expect(trafficPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
			Expect(accessPolicy.Status.Workloads).To(HaveLen(1))
			Expect(accessPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
		})
		It("sets policy workloads using virtual mesh", func() {
			snap := input.NewInputSnapshotManualBuilder("").
				AddTrafficPolicies(v1alpha2.TrafficPolicySlice{trafficPolicy}).
				AddAccessPolicies(v1alpha2.AccessPolicySlice{accessPolicy}).
				AddTrafficTargets(discoveryv1alpha2.TrafficTargetSlice{trafficTarget}).
				AddWorkloads(discoveryv1alpha2.WorkloadSlice{workload1, workload2}).
				AddMeshes(discoveryv1alpha2.MeshSlice{mesh1, mesh2}).
				AddVirtualMeshes(v1alpha2.VirtualMeshSlice{virtualMesh}).
				Build()
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap)

			// trafficTarget is in mesh1, workload1 is in mesh1, and workload2 is in mesh2.
			// since mesh1 and mesh2 are in the same virtual mesh, both workloads are returned
			Expect(trafficPolicy.Status.Workloads).To(HaveLen(2))
			Expect(trafficPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
			Expect(trafficPolicy.Status.Workloads[1]).To(Equal(sets.Key(workload2)))
			Expect(accessPolicy.Status.Workloads).To(HaveLen(2))
			Expect(accessPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
			Expect(accessPolicy.Status.Workloads[1]).To(Equal(sets.Key(workload2)))
		})
		It("sets no policy workloads when there is no matching mesh", func() {
			workload1.Spec.Mesh.Name = "mesh2"
			snap := input.NewInputSnapshotManualBuilder("").
				AddTrafficPolicies(v1alpha2.TrafficPolicySlice{trafficPolicy}).
				AddAccessPolicies(v1alpha2.AccessPolicySlice{accessPolicy}).
				AddTrafficTargets(discoveryv1alpha2.TrafficTargetSlice{trafficTarget}).
				AddWorkloads(discoveryv1alpha2.WorkloadSlice{workload1, workload2}).
				AddMeshes(discoveryv1alpha2.MeshSlice{mesh1, mesh2}).
				Build()
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap)

			// trafficTarget is in mesh1, but both workloads are in mesh2
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

func (t testIstioTranslator) Translate(ctx context.Context, in input.Snapshot, reporter reporting.Reporter) (translation.OutputSnapshots, error) {
	t.callReporter(reporter)
	return translation.OutputSnapshots{}, nil
}
