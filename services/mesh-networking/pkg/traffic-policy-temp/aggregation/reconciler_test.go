package traffic_policy_aggregation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation"
	mock_traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/mocks"
	mock_zephyr_discovery_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Traffic Policy Aggregation Reconciler", func() {
	var (
		ctx  = context.TODO()
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("does not pick up invalid or not-yet-validated traffic policies", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)

		tps := []*zephyr_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_INVALID,
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_UNKNOWN,
					},
				},
			},
		}

		trafficPolicyClient.EXPECT().
			ListTrafficPolicy(ctx).
			Return(&zephyr_networking.TrafficPolicyList{
				Items: []zephyr_networking.TrafficPolicy{*tps[0], *tps[1]},
			}, nil)
		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&zephyr_discovery.MeshServiceList{}, nil)
		aggregator.EXPECT().
			GroupByMeshService(nil, map[*zephyr_discovery.MeshService]string{}).
			Return([]*traffic_policy_aggregation.ServiceWithRelevantPolicies{})

		reconciler := traffic_policy_aggregation.NewAggregationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			aggregator,
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can aggregate traffic policies", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)

		// going to associate tp1 and tp2 with service 1; and tp3 and tp4 with service 2
		tps := []*zephyr_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
				},
			},
		}
		cluster1 := "cluster-1"
		cluster2 := "cluster-2"
		meshServices := []*zephyr_discovery.MeshService{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:   "ms1",
					Labels: map[string]string{constants.COMPUTE_TARGET: cluster1},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:   "ms2",
					Labels: map[string]string{constants.COMPUTE_TARGET: cluster2},
				},
			},
		}

		trafficPolicyClient.EXPECT().
			ListTrafficPolicy(ctx).
			Return(&zephyr_networking.TrafficPolicyList{
				Items: []zephyr_networking.TrafficPolicy{*tps[0], *tps[1], *tps[2], *tps[3]},
			}, nil)
		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&zephyr_discovery.MeshServiceList{
				Items: []zephyr_discovery.MeshService{*meshServices[0], *meshServices[1]},
			}, nil)
		aggregator.EXPECT().
			GroupByMeshService(tps, gomock.Any()).
			DoAndReturn(func(
				trafficPolicies []*zephyr_networking.TrafficPolicy,
				meshServiceToClusterName map[*zephyr_discovery.MeshService]string,
			) []*traffic_policy_aggregation.ServiceWithRelevantPolicies {
				// see https://github.com/solo-io/service-mesh-hub/issues/677 for why this is such a pain
				Expect(meshServiceToClusterName).To(HaveLen(2))
				for service, cluster := range meshServiceToClusterName {
					properAssociation := (service.ObjectMeta.Name == "ms1" && cluster == cluster1) ||
						(service.ObjectMeta.Name == "ms2" && cluster == cluster2)
					Expect(properAssociation).To(BeTrue())
				}
				return []*traffic_policy_aggregation.ServiceWithRelevantPolicies{
					{
						MeshService:     meshServices[0],
						TrafficPolicies: tps[0:2],
					},
					{
						MeshService:     meshServices[1],
						TrafficPolicies: tps[2:4],
					},
				}
			})
		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:   "ms1",
					Labels: map[string]string{constants.COMPUTE_TARGET: cluster1},
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Name:              "tp1",
							TrafficPolicySpec: &tps[0].Spec,
						},
						{
							Name:              "tp2",
							TrafficPolicySpec: &tps[1].Spec,
						},
					},
				},
			}).
			Return(nil)
		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:   "ms2",
					Labels: map[string]string{constants.COMPUTE_TARGET: cluster2},
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Name:              "tp3",
							TrafficPolicySpec: &tps[2].Spec,
						},
						{
							Name:              "tp4",
							TrafficPolicySpec: &tps[3].Spec,
						},
					},
				},
			}).
			Return(nil)

		reconciler := traffic_policy_aggregation.NewAggregationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			aggregator,
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does not replace updated policies if they cannot merge with the existing state", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)

		// going to associate tp1 and tp2 with service 1; and tp3 and tp4 with service 2
		tps := []*zephyr_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
				Spec: types.TrafficPolicySpec{
					Retries: &types.TrafficPolicySpec_RetryPolicy{
						Attempts: 1,
					},
				},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
				},
			},
			{
				// This traffic policy will be in conflict with the existing state on the service
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp2"},
				Spec: types.TrafficPolicySpec{
					Retries: &types.TrafficPolicySpec_RetryPolicy{
						Attempts: 9999,
					},
				},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp3"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp4"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
				},
			},
		}
		tp2LastValidState := *tps[1]
		tp2LastValidState.Spec = types.TrafficPolicySpec{
			Retries: &types.TrafficPolicySpec_RetryPolicy{
				Attempts: 1,
			},
		}

		cluster1 := "cluster-1"
		cluster2 := "cluster-2"
		meshServices := []*zephyr_discovery.MeshService{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:   "ms1",
					Labels: map[string]string{constants.COMPUTE_TARGET: cluster1},
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Name:              tp2LastValidState.Name,
							TrafficPolicySpec: &tp2LastValidState.Spec,
						},
						{
							Name:              tps[0].Name,
							TrafficPolicySpec: &tps[0].Spec,
						},
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:   "ms2",
					Labels: map[string]string{constants.COMPUTE_TARGET: cluster2},
				},
			},
		}
		conflictError := &types.TrafficPolicyStatus_ConflictError{
			ErrorMessage: "whoops conflict",
		}

		trafficPolicyClient.EXPECT().
			ListTrafficPolicy(ctx).
			Return(&zephyr_networking.TrafficPolicyList{
				Items: []zephyr_networking.TrafficPolicy{*tps[0], *tps[1], *tps[2], *tps[3]},
			}, nil)
		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&zephyr_discovery.MeshServiceList{
				Items: []zephyr_discovery.MeshService{*meshServices[0], *meshServices[1]},
			}, nil)
		aggregator.EXPECT().
			GroupByMeshService(tps, gomock.Any()).
			DoAndReturn(func(
				trafficPolicies []*zephyr_networking.TrafficPolicy,
				meshServiceToClusterName map[*zephyr_discovery.MeshService]string,
			) []*traffic_policy_aggregation.ServiceWithRelevantPolicies {
				// see https://github.com/solo-io/service-mesh-hub/issues/677 for why this is such a pain
				Expect(meshServiceToClusterName).To(HaveLen(2))
				for service, cluster := range meshServiceToClusterName {
					properAssociation := (service.ObjectMeta.Name == "ms1" && cluster == cluster1) ||
						(service.ObjectMeta.Name == "ms2" && cluster == cluster2)
					Expect(properAssociation).To(BeTrue())
				}
				return []*traffic_policy_aggregation.ServiceWithRelevantPolicies{
					{
						MeshService:     meshServices[0],
						TrafficPolicies: tps[0:2],
					},
					{
						MeshService:     meshServices[1],
						TrafficPolicies: tps[2:4],
					},
				}
			})
		aggregator.EXPECT().
			FindMergeConflict(&tps[1].Spec, []*types.TrafficPolicySpec{
				&tps[0].Spec,
			}, meshServices).
			Return(conflictError)

		// NOTE: No update issued for ms1

		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:   "ms2",
					Labels: map[string]string{constants.COMPUTE_TARGET: cluster2},
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Name:              "tp3",
							TrafficPolicySpec: &tps[2].Spec,
						},
						{
							Name:              "tp4",
							TrafficPolicySpec: &tps[3].Spec,
						},
					},
				},
			}).
			Return(nil)
		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &zephyr_networking.TrafficPolicy{
				ObjectMeta: tps[1].ObjectMeta,
				Spec:       tps[1].Spec,
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
					ConflictErrors: []*types.TrafficPolicyStatus_ConflictError{conflictError},
				},
			})

		reconciler := traffic_policy_aggregation.NewAggregationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			aggregator,
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})
})
