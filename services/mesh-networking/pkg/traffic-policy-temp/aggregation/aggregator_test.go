package traffic_policy_aggregation_test

import (
	types2 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Aggregator", func() {
	var (
		ctrl            *gomock.Controller
		trafficPolicies = []*zephyr_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp-1"},
				Spec: types.TrafficPolicySpec{
					DestinationSelector: &zephyr_core_types.ServiceSelector{
						ServiceSelectorType: &zephyr_core_types.ServiceSelector_Matcher_{},
					},
					Retries: &types.TrafficPolicySpec_RetryPolicy{
						Attempts: 1,
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp-2"},
				Spec: types.TrafficPolicySpec{
					DestinationSelector: &zephyr_core_types.ServiceSelector{
						ServiceSelectorType: &zephyr_core_types.ServiceSelector_ServiceRefs_{},
					},
					Retries: &types.TrafficPolicySpec_RetryPolicy{
						Attempts: 2,
					},
				},
			},
		}
		meshService = &zephyr_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: "ms-1"},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Grouping", func() {
		It("returns empty when no traffic policies are provided", func() {
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			aggregator := traffic_policy_aggregation.NewAggregator(resourceSelector)

			policies, err := aggregator.PoliciesForService([]*zephyr_networking.TrafficPolicy{}, meshService)
			Expect(err).NotTo(HaveOccurred())
			Expect(policies).To(BeEmpty())
		})

		It("can associate policies with services", func() {
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			aggregator := traffic_policy_aggregation.NewAggregator(resourceSelector)

			resourceSelector.EXPECT().
				FilterMeshServicesByServiceSelector(
					[]*zephyr_discovery.MeshService{meshService},
					trafficPolicies[0].Spec.DestinationSelector,
				).
				Return([]*zephyr_discovery.MeshService{meshService}, nil)
			resourceSelector.EXPECT().
				FilterMeshServicesByServiceSelector(
					[]*zephyr_discovery.MeshService{meshService},
					trafficPolicies[1].Spec.DestinationSelector,
				).
				Return([]*zephyr_discovery.MeshService{meshService}, nil)

			policies, err := aggregator.PoliciesForService(trafficPolicies, meshService)
			Expect(err).NotTo(HaveOccurred())
			Expect(policies).To(Equal(trafficPolicies))
		})

		It("can associate some but not all policies with a service", func() {
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			aggregator := traffic_policy_aggregation.NewAggregator(resourceSelector)

			resourceSelector.EXPECT().
				FilterMeshServicesByServiceSelector(
					[]*zephyr_discovery.MeshService{meshService},
					trafficPolicies[0].Spec.DestinationSelector,
				).
				Return([]*zephyr_discovery.MeshService{meshService}, nil)
			resourceSelector.EXPECT().
				FilterMeshServicesByServiceSelector(
					[]*zephyr_discovery.MeshService{meshService},
					trafficPolicies[1].Spec.DestinationSelector,
				).
				Return(nil, nil)

			policies, err := aggregator.PoliciesForService(trafficPolicies, meshService)
			Expect(err).NotTo(HaveOccurred())
			Expect(policies).To(Equal(trafficPolicies[:1]))
		})
	})

	Context("Merge conflicts", func() {
		It("reports an error for policies that conflict", func() {
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			aggregator := traffic_policy_aggregation.NewAggregator(resourceSelector)

			var specs []*types.TrafficPolicySpec
			for index, tpIter := range trafficPolicies {
				tp := tpIter
				if index == 0 {
					continue
				}

				specs = append(specs, &tp.Spec)
			}
			conflictError := aggregator.FindMergeConflict(&trafficPolicies[0].Spec, specs, meshService)
			Expect(conflictError).NotTo(BeNil())
			Expect(conflictError.ErrorMessage).To(Equal(traffic_policy_aggregation.TrafficPolicyConflictError.Error()))
		})

		It("does not report an error when no conflict", func() {
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			aggregator := traffic_policy_aggregation.NewAggregator(resourceSelector)

			conflictError := aggregator.FindMergeConflict(&trafficPolicies[0].Spec, []*types.TrafficPolicySpec{
				{
					Retries: &types.TrafficPolicySpec_RetryPolicy{
						Attempts: trafficPolicies[0].Spec.Retries.Attempts,
					},
					RequestTimeout: &types2.Duration{
						Seconds: 420,
					},
				},
			}, meshService)
			Expect(conflictError).To(BeNil())
		})
	})
})
