package traffic_policy_validation_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	v1alpha12 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/kube"
	"github.com/solo-io/service-mesh-hub/pkg/kube/selection"
	traffic_policy_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/validation"
	mock_traffic_policy_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/validation/mocks"
	mock_zephyr_discovery_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Validation Reconciler", func() {
	var (
		ctx                 = context.TODO()
		ctrl                *gomock.Controller
		trafficPolicyClient *mock_zephyr_networking_clients.MockTrafficPolicyClient
		meshServiceClient   *mock_zephyr_discovery_clients.MockMeshServiceClient
		validator           *mock_traffic_policy_validation.MockValidator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		trafficPolicyClient = mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient = mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		validator = mock_traffic_policy_validation.NewMockValidator(ctrl)

	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can set a new validation status", func() {
		invalidTrafficPolicy := &v1alpha12.TrafficPolicy{}
		failedValidationStatus := &zephyr_core_types.Status{
			State: zephyr_core_types.Status_INVALID,
		}

		trafficPolicyClient.EXPECT().
			ListTrafficPolicy(ctx).
			Return(&v1alpha12.TrafficPolicyList{
				Items: []v1alpha12.TrafficPolicy{*invalidTrafficPolicy},
			}, nil)

		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&v1alpha1.MeshServiceList{}, nil)

		validator.EXPECT().
			ValidateTrafficPolicy(invalidTrafficPolicy, nil).
			Return(failedValidationStatus, nil)

		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &v1alpha12.TrafficPolicy{
				Status: types.TrafficPolicyStatus{
					ValidationStatus: failedValidationStatus,
				},
			})

		reconciler := traffic_policy_validation.NewValidationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			validator,
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does not issue an update if the status is up-to-date", func() {
		failedValidationStatus := &zephyr_core_types.Status{
			State: zephyr_core_types.Status_INVALID,
		}
		alreadyInvalidTrafficPolicy := &v1alpha12.TrafficPolicy{
			Status: types.TrafficPolicyStatus{
				ValidationStatus: failedValidationStatus,
			},
		}

		trafficPolicyClient.EXPECT().
			ListTrafficPolicy(ctx).
			Return(&v1alpha12.TrafficPolicyList{
				Items: []v1alpha12.TrafficPolicy{*alreadyInvalidTrafficPolicy},
			}, nil)

		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&v1alpha1.MeshServiceList{}, nil)

		validator.EXPECT().
			ValidateTrafficPolicy(alreadyInvalidTrafficPolicy, nil).
			Return(failedValidationStatus, nil)

		reconciler := traffic_policy_validation.NewValidationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			validator,
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does issue an update if the generation is not up-to-date", func() {
		status := &zephyr_core_types.Status{
			State: zephyr_core_types.Status_INVALID,
		}
		trafficPolicy := v1alpha12.TrafficPolicy{
			ObjectMeta: v1.ObjectMeta{Generation: 2},
			Status: types.TrafficPolicyStatus{
				ValidationStatus:   status,
				ObservedGeneration: 1,
			},
		}

		trafficPolicyClient.EXPECT().
			ListTrafficPolicy(ctx).
			Return(&v1alpha12.TrafficPolicyList{
				Items: []v1alpha12.TrafficPolicy{trafficPolicy},
			}, nil)

		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&v1alpha1.MeshServiceList{}, nil)

		validator.EXPECT().
			ValidateTrafficPolicy(&trafficPolicy, nil).
			Return(status, nil)

		updatedTrafficPolicy := trafficPolicy
		updatedTrafficPolicy.Status.ObservedGeneration = trafficPolicy.Generation

		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &updatedTrafficPolicy)

		reconciler := traffic_policy_validation.NewValidationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			validator,
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("benchmarks", func() {

		Measure("it reconciles traffic policies", func(b Benchmarker) {
			// not using mock client, as we don't want to measure their (lack of) overhead
			var tp trafficPolicyBenchmarkClient
			var ms meshServiceBenchmarkClient

			const trafficPolicies = 1000
			const meshServices = 1000

			for i := 0; i < trafficPolicies; i++ {
				tp.policies.Items = append(tp.policies.Items, v1alpha12.TrafficPolicy{
					ObjectMeta: v1.ObjectMeta{Name: fmt.Sprintf("tp-%d", i)},
					Spec: types.TrafficPolicySpec{
						SourceSelector: &zephyr_core_types.WorkloadSelector{
							Labels: map[string]string{
								"foo": "bar",
							},
						},
						TrafficShift: &types.TrafficPolicySpec_MultiDestination{
							Destinations: []*types.TrafficPolicySpec_MultiDestination_WeightedDestination{
								{
									Destination: &zephyr_core_types.ResourceRef{
										Name:      "reviews",
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
				ms.services.Items = append(ms.services.Items, v1alpha1.MeshService{
					ObjectMeta: v1.ObjectMeta{
						Name: fmt.Sprintf("sm-%d", i),
						Labels: map[string]string{
							"foo":                       "bar",
							kube.KUBE_SERVICE_NAME:      "reviews",
							kube.KUBE_SERVICE_NAMESPACE: "reviews",
							kube.COMPUTE_TARGET:         "test",
						},
					},
					Spec: discovery_types.MeshServiceSpec{},
				})
			}

			validator := traffic_policy_validation.NewValidator(selection.NewBaseResourceSelector())

			reconciler := traffic_policy_validation.NewValidationReconciler(&tp, &ms, validator)
			ctx := context.Background()
			runtime := b.Time("runtime", func() {
				//	output := SomethingHard()
				//	Expect(output).To(Equal(17))
				reconciler.Reconcile(ctx)
			})

			Ω(runtime.Seconds()).Should(BeNumerically("<", 0.01), "validator.Reconcile() shouldn't take too long.")

			//	b.RecordValue("disk usage (in MB)", HowMuchDiskSpaceDidYouUse())
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
