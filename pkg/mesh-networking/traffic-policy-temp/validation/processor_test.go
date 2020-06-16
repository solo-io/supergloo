package traffic_policy_validation_test

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	v1alpha12 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	traffic_policy_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/validation"
	mock_traffic_policy_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/validation/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Validation Reconciler", func() {
	var (
		ctx       = context.TODO()
		ctrl      *gomock.Controller
		validator *mock_traffic_policy_validation.MockValidator
		processor traffic_policy_validation.ValidationProcessor
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		validator = mock_traffic_policy_validation.NewMockValidator(ctrl)
		processor = traffic_policy_validation.NewValidationProcessor(validator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can set a new validation status", func() {
		invalidTrafficPolicy := &v1alpha12.TrafficPolicy{}
		failedValidationStatus := &smh_core_types.Status{
			State: smh_core_types.Status_INVALID,
		}

		validator.EXPECT().
			ValidateTrafficPolicy(invalidTrafficPolicy, nil).
			Return(failedValidationStatus, nil)

		updatedPolicies := processor.Process(ctx, []*v1alpha12.TrafficPolicy{invalidTrafficPolicy}, nil)

		Expect(updatedPolicies).To(HaveLen(1))
		// make sure the same pointer is returned; i.e. don't use BeEquivalent
		Expect(updatedPolicies[0]).To(BeIdenticalTo(invalidTrafficPolicy))
	})

	It("does not issue an update if the status is up-to-date", func() {
		failedValidationStatus := &smh_core_types.Status{
			State: smh_core_types.Status_INVALID,
		}
		alreadyInvalidTrafficPolicy := &v1alpha12.TrafficPolicy{
			Status: types.TrafficPolicyStatus{
				ValidationStatus: failedValidationStatus,
			},
		}

		validator.EXPECT().
			ValidateTrafficPolicy(alreadyInvalidTrafficPolicy, nil).
			Return(failedValidationStatus, nil)

		updatedPolicies := processor.Process(ctx, []*v1alpha12.TrafficPolicy{alreadyInvalidTrafficPolicy}, nil)

		Expect(updatedPolicies).To(BeEmpty())
	})

	It("does issue an update if the generation is not up-to-date", func() {
		status := &smh_core_types.Status{
			State: smh_core_types.Status_INVALID,
		}
		trafficPolicy := &v1alpha12.TrafficPolicy{
			ObjectMeta: v1.ObjectMeta{Generation: 2},
			Status: types.TrafficPolicyStatus{
				ValidationStatus:   status,
				ObservedGeneration: 1,
			},
		}

		validator.EXPECT().
			ValidateTrafficPolicy(trafficPolicy, nil).
			Return(status, nil)

		updatedPolicies := processor.Process(ctx, []*v1alpha12.TrafficPolicy{trafficPolicy}, nil)

		Expect(updatedPolicies).To(HaveLen(1))
		// make sure the same pointer is returned; i.e. don't use BeEquivalent
		Expect(updatedPolicies[0]).To(BeIdenticalTo(trafficPolicy))
		Expect(trafficPolicy.Generation).To(Equal(trafficPolicy.Status.ObservedGeneration))
	})

	Context("benchmarks", func() {

		Measure("it reconciles traffic policies", func(b Benchmarker) {
			// not using mock client, as we don't want to measure their (lack of) overhead
			var tp []*v1alpha12.TrafficPolicy
			var ms []*v1alpha1.MeshService

			const trafficPolicies = 1000
			const meshServices = 1000

			for i := 0; i < trafficPolicies; i++ {
				tp = append(tp, &v1alpha12.TrafficPolicy{
					ObjectMeta: v1.ObjectMeta{Name: fmt.Sprintf("tp-%d", i)},
					Spec: types.TrafficPolicySpec{
						SourceSelector: &smh_core_types.WorkloadSelector{
							Labels: map[string]string{
								"foo": "bar",
							},
						},
						TrafficShift: &types.TrafficPolicySpec_MultiDestination{
							Destinations: []*types.TrafficPolicySpec_MultiDestination_WeightedDestination{
								{
									Destination: &smh_core_types.ResourceRef{
										Name:      fmt.Sprintf("reviews-%d", i),
										Namespace: "reviews",
										Cluster:   "test",
									},
									Weight: 100,
								},
							},
						},
					},
				})
			}
			for i := 0; i < meshServices; i++ {
				ms = append(ms, &v1alpha1.MeshService{
					ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("sm-%d", i),
						Labels: map[string]string{
							"foo":                       "bar",
							kube.KUBE_SERVICE_NAME:      fmt.Sprintf("reviews-%d", i),
							kube.KUBE_SERVICE_NAMESPACE: "reviews",
							kube.COMPUTE_TARGET:         "test",
						},
					},
					Spec: discovery_types.MeshServiceSpec{},
				})
			}
			// suffle things for more realistic test
			rand.Shuffle(len(ms), func(i, j int) {
				item := ms[j]
				ms[j] = ms[i]
				ms[i] = item
			})

			validator := traffic_policy_validation.NewValidator(selection.NewBaseResourceSelector())

			processor := traffic_policy_validation.NewValidationProcessor(validator)
			ctx := context.Background()
			runtime := b.Time("runtime", func() {
				processor.Process(ctx, tp, ms)
			})
			// ideally should be less than 1ms; but 1s is good for now. in practice it's around 10ms.
			Ω(runtime.Seconds()).Should(BeNumerically("<", 1), "validator.Reconcile() shouldn't take too long.")
		}, 10)

	})
})

type trafficPolicyBenchmarkClient struct {
	policies v1alpha12.TrafficPolicyList
}

func (t *trafficPolicyBenchmarkClient) GetTrafficPolicy(ctx context.Context, key client.ObjectKey) (*v1alpha12.TrafficPolicy, error) {
	panic("not implemented")
}
func (t *trafficPolicyBenchmarkClient) ListTrafficPolicy(ctx context.Context, opts ...client.ListOption) (*v1alpha12.TrafficPolicyList, error) {
	return &t.policies, nil
}
func (t *trafficPolicyBenchmarkClient) UpdateTrafficPolicyStatus(ctx context.Context, obj *v1alpha12.TrafficPolicy, opts ...client.UpdateOption) error {
	return nil
}
func (t *trafficPolicyBenchmarkClient) PatchTrafficPolicyStatus(ctx context.Context, obj *v1alpha12.TrafficPolicy, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

type meshServiceBenchmarkClient struct {
	services v1alpha1.MeshServiceList
}

func (t *meshServiceBenchmarkClient) GetMeshService(ctx context.Context, key client.ObjectKey) (*v1alpha1.MeshService, error) {
	panic("not implemented")
}

func (t *meshServiceBenchmarkClient) ListMeshService(ctx context.Context, opts ...client.ListOption) (*v1alpha1.MeshServiceList, error) {
	return &t.services, nil
}
