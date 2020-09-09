package apply_test

import (
	"context"
	"errors"

	corev1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation"
	corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	v1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
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
			}
			workload = &discoveryv1alpha2.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld1",
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

			snap = input.NewSnapshot("",
				discoveryv1alpha2sets.NewTrafficTargetSet(discoveryv1alpha2.TrafficTargetSlice{
					trafficTarget,
				}...),
				discoveryv1alpha2sets.NewWorkloadSet(discoveryv1alpha2.WorkloadSlice{
					workload,
				}...),
				discoveryv1alpha2sets.NewMeshSet(),
				v1alpha2sets.NewTrafficPolicySet(v1alpha2.TrafficPolicySlice{
					trafficPolicy1,
					trafficPolicy2,
				}...),
				v1alpha2sets.NewAccessPolicySet(),
				v1alpha2sets.NewVirtualMeshSet(),
				v1alpha2sets.NewFailoverServiceSet(),
				corev1sets.NewSecretSet(),
				skv1alpha1sets.NewKubernetesClusterSet(),
			)
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

			snap = input.NewSnapshot("",
				discoveryv1alpha2sets.NewTrafficTargetSet(discoveryv1alpha2.TrafficTargetSlice{
					trafficTarget,
				}...),
				discoveryv1alpha2sets.NewWorkloadSet(),
				discoveryv1alpha2sets.NewMeshSet(),
				v1alpha2sets.NewTrafficPolicySet(v1alpha2.TrafficPolicySlice{
					trafficPolicy,
				}...),
				v1alpha2sets.NewAccessPolicySet(),
				v1alpha2sets.NewVirtualMeshSet(),
				v1alpha2sets.NewFailoverServiceSet(),
				corev1sets.NewSecretSet(),
				skv1alpha1sets.NewKubernetesClusterSet(),
			)
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

	Context("validate one VirtualMesh per mesh", func() {
		var (
			vm1 = &v1alpha2.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "namespace1",
				},
				Spec: v1alpha2.VirtualMeshSpec{
					Meshes: []*corev1.ObjectRef{
						{
							Name:      "mesh1",
							Namespace: "namespace1",
						},
					},
				},
				Status: v1alpha2.VirtualMeshStatus{
					ObservedGeneration: 0,
					State:              v1alpha2.ApprovalState_ACCEPTED,
				},
			}
			vm2 = &v1alpha2.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm2",
					Namespace: "namespace1",
				},
				Spec: v1alpha2.VirtualMeshSpec{
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
			}
			vm3 = &v1alpha2.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm3",
					Namespace: "namespace1",
				},
				Spec: v1alpha2.VirtualMeshSpec{
					Meshes: []*corev1.ObjectRef{
						{
							Name:      "mesh2",
							Namespace: "namespace1",
						},
					},
				},
			}
			vm4 = &v1alpha2.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm4",
					Namespace: "namespace1",
				},
				Spec: v1alpha2.VirtualMeshSpec{
					Meshes: []*corev1.ObjectRef{
						{
							Name:      "mesh2",
							Namespace: "namespace1",
						},
					},
				},
			}
			vm5 = &v1alpha2.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm5",
					Namespace: "namespace1",
				},
				Spec: v1alpha2.VirtualMeshSpec{
					Meshes: []*corev1.ObjectRef{
						{
							Name:      "mesh3",
							Namespace: "namespace1",
						},
					},
				},
			}
			snap = input.NewSnapshot("",
				discoveryv1alpha2sets.NewTrafficTargetSet(),
				discoveryv1alpha2sets.NewWorkloadSet(),
				discoveryv1alpha2sets.NewMeshSet(),
				v1alpha2sets.NewTrafficPolicySet(),
				v1alpha2sets.NewAccessPolicySet(),
				v1alpha2sets.NewVirtualMeshSet(
					vm1, vm2, vm3, vm4, vm5,
				),
				v1alpha2sets.NewFailoverServiceSet(),
				corev1sets.NewSecretSet(),
				skv1alpha1sets.NewKubernetesClusterSet(),
			)
		)
		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				return
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap)
		})
		It("vm1 should render vm2 invalid, which should permit vm3", func() {
			Expect(vm2.Status.State).To(Equal(v1alpha2.ApprovalState_INVALID))
			Expect(vm3.Status.State).To(Equal(v1alpha2.ApprovalState_ACCEPTED))
		})
		It("vm5 should be permitted given that vm4 is invalid", func() {
			Expect(vm5.Status.State).To(Equal(v1alpha2.ApprovalState_ACCEPTED))
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
