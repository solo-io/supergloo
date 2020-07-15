package validation_test

import (
	"context"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discoveryv1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/output/istio"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/smh/pkg/mesh-networking/reporter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/smh/pkg/mesh-networking/validation"
)

var _ = Describe("Validator", func() {
	Context("valid traffic policies", func() {
		var (
			meshService = &discoveryv1alpha1.MeshService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
			}
			trafficPolicy1 = &v1alpha1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
				Spec: v1alpha1.TrafficPolicySpec{
					// fill an arbitrary part of the spec
					Mirror: &v1alpha1.TrafficPolicySpec_Mirror{},
				},
			}
			trafficPolicy2 = &v1alpha1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp2",
					Namespace: "ns",
				},
				Spec: v1alpha1.TrafficPolicySpec{
					// fill an arbitrary part of the spec
					FaultInjection: &v1alpha1.TrafficPolicySpec_FaultInjection{},
				},
			}

			snap = input.NewSnapshot("",
				discoveryv1alpha1sets.NewMeshServiceSet(discoveryv1alpha1.MeshServiceSlice{
					meshService,
				}...),
				discoveryv1alpha1sets.NewMeshWorkloadSet(),
				discoveryv1alpha1sets.NewMeshSet(),
				v1alpha1sets.NewTrafficPolicySet(v1alpha1.TrafficPolicySlice{
					trafficPolicy1,
					trafficPolicy2,
				}...),
				v1alpha1sets.NewAccessPolicySet(),
				v1alpha1sets.NewVirtualMeshSet(),
				skv1alpha1sets.NewKubernetesClusterSet(),
			)
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporter.Reporter) {
				// no report = accept
			}}
			validator := NewValidator(translator)
			validator.Validate(context.TODO(), snap)

		})
		It("updates status on input traffic policies", func() {
			Expect(trafficPolicy1.Status.MeshServices).To(HaveKey(sets.Key(meshService)))
			Expect(trafficPolicy1.Status.MeshServices[sets.Key(meshService)]).To(Equal(&v1alpha1.ValidationStatus{
				AcceptanceOrder: 0,
				State:           v1alpha1.ValidationState_ACCEPTED,
			}))
			Expect(trafficPolicy2.Status.MeshServices).To(HaveKey(sets.Key(meshService)))
			Expect(trafficPolicy2.Status.MeshServices[sets.Key(meshService)]).To(Equal(&v1alpha1.ValidationStatus{
				AcceptanceOrder: 1,
				State:           v1alpha1.ValidationState_ACCEPTED,
			}))

		})
		It("updates status on input mesh services policies", func() {
			Expect(meshService.Status.AppliedTrafficPolicies).To(HaveLen(2))
			Expect(meshService.Status.AppliedTrafficPolicies[0].Ref).To(Equal(ezkube.MakeObjectRef(trafficPolicy1)))
			Expect(meshService.Status.AppliedTrafficPolicies[0].Spec).To(Equal(&trafficPolicy1.Spec))
			Expect(meshService.Status.AppliedTrafficPolicies[1].Ref).To(Equal(ezkube.MakeObjectRef(trafficPolicy2)))
			Expect(meshService.Status.AppliedTrafficPolicies[1].Spec).To(Equal(&trafficPolicy2.Spec))
		})
	})
	Context("invalid traffic policies", func() {
		var (
			meshService = &discoveryv1alpha1.MeshService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
			}
			trafficPolicy = &v1alpha1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
			}

			snap = input.NewSnapshot("",
				discoveryv1alpha1sets.NewMeshServiceSet(discoveryv1alpha1.MeshServiceSlice{
					meshService,
				}...),
				discoveryv1alpha1sets.NewMeshWorkloadSet(),
				discoveryv1alpha1sets.NewMeshSet(),
				v1alpha1sets.NewTrafficPolicySet(v1alpha1.TrafficPolicySlice{
					trafficPolicy,
				}...),
				v1alpha1sets.NewAccessPolicySet(),
				v1alpha1sets.NewVirtualMeshSet(),
				skv1alpha1sets.NewKubernetesClusterSet(),
			)
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporter.Reporter) {
				// report = reject
				reporter.ReportTrafficPolicy(meshService, trafficPolicy, errors.New("did an oopsie"))
			}}
			validator := NewValidator(translator)
			validator.Validate(context.TODO(), snap)

		})
		It("updates status on input traffic policies", func() {
			Expect(trafficPolicy.Status.MeshServices).To(HaveKey(sets.Key(meshService)))
			Expect(trafficPolicy.Status.MeshServices[sets.Key(meshService)]).To(Equal(&v1alpha1.ValidationStatus{
				AcceptanceOrder: 0,
				State:           v1alpha1.ValidationState_INVALID,
				Errors:          []string{"did an oopsie"},
			}))
		})
		It("does not add the policy to the mesh service status", func() {
			Expect(meshService.Status.AppliedTrafficPolicies).To(HaveLen(0))
		})
	})
})

// NOTE(ilackarms): we implement a test translator here instead of using a mock because
// we need to call methods on the reporter which is passed as an argument to the translator
type testIstioTranslator struct {
	callReporter func(reporter reporter.Reporter)
}

func (t testIstioTranslator) Translate(
	in input.Snapshot,
	reporter reporter.Reporter,
) (istio.Snapshot, error) {
	t.callReporter(reporter)
	return nil, nil
}
