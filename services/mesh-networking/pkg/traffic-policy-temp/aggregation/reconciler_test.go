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
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation"
	mock_traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/mocks"
	mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/meshes"
	mock_mesh_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/meshes/mocks"
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
		meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
		validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)

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
			GroupByMeshService(nil, nil).
			Return([]*traffic_policy_aggregation.ServiceWithRelevantPolicies{}, nil)

		reconciler := traffic_policy_aggregation.NewAggregationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			meshClient,
			aggregator,
			map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator{
				zephyr_core_types.MeshType_ISTIO: validator,
			},
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can aggregate traffic policies", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
		meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
		validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)

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
		mesh1 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-1",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster1,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}
		mesh2 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-2",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster2,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}

		meshServices := []*zephyr_discovery.MeshService{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms1",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh1.ObjectMeta),
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms2",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh2.ObjectMeta),
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
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh1.ObjectMeta)).
			Return(mesh1, nil)
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh2.ObjectMeta)).
			Return(mesh2, nil)
		aggregator.EXPECT().
			GroupByMeshService(tps, gomock.Any()).
			DoAndReturn(func(
				trafficPolicies []*zephyr_networking.TrafficPolicy,
				meshServices []*zephyr_discovery.MeshService,
			) (result []*traffic_policy_aggregation.ServiceWithRelevantPolicies, err error) {
				// see https://github.com/solo-io/service-mesh-hub/issues/677 for why this is such a pain
				Expect(meshServices).To(HaveLen(2))
				var ret []*traffic_policy_aggregation.ServiceWithRelevantPolicies
				for _, service := range meshServices {
					var policies []*zephyr_networking.TrafficPolicy
					if service.ObjectMeta.Name == "ms1" {
						policies = trafficPolicies[0:2]
					} else {
						policies = trafficPolicies[2:4]
					}
					ret = append(ret, &traffic_policy_aggregation.ServiceWithRelevantPolicies{
						MeshService:     service,
						TrafficPolicies: policies,
					})
				}

				return ret, nil
			})
		aggregator.EXPECT().
			FindMergeConflict(&tps[0].Spec, nil, meshServices[0]).
			Return(nil)
		aggregator.EXPECT().
			FindMergeConflict(&tps[1].Spec, []*types.TrafficPolicySpec{&tps[0].Spec}, meshServices[0]).
			Return(nil)
		aggregator.EXPECT().
			FindMergeConflict(&tps[2].Spec, nil, meshServices[1]).
			Return(nil)
		aggregator.EXPECT().
			FindMergeConflict(&tps[3].Spec, []*types.TrafficPolicySpec{&tps[2].Spec}, meshServices[1]).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[0], mesh1, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{{
				Ref: &zephyr_core_types.ResourceRef{
					Name: "tp1",
				},
				TrafficPolicySpec: &tps[0].Spec,
			}}).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[0], mesh1, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: "tp1",
					},
					TrafficPolicySpec: &tps[0].Spec,
				},
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: "tp2",
					},
					TrafficPolicySpec: &tps[1].Spec,
				},
			}).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[1], mesh2, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{{
				Ref: &zephyr_core_types.ResourceRef{
					Name: "tp3",
				},
				TrafficPolicySpec: &tps[2].Spec,
			}}).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[1], mesh2, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: "tp3",
					},
					TrafficPolicySpec: &tps[1].Spec,
				},
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: "tp4",
					},
					TrafficPolicySpec: &tps[2].Spec,
				},
			}).
			Return(nil)

		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms1",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh1.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp1",
							},
							TrafficPolicySpec: &tps[0].Spec,
						},
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp2",
							},
							TrafficPolicySpec: &tps[1].Spec,
						},
					},
				},
			}).
			Return(nil)
		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms2",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh2.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp3",
							},
							TrafficPolicySpec: &tps[2].Spec,
						},
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp4",
							},
							TrafficPolicySpec: &tps[3].Spec,
						},
					},
				},
			}).
			Return(nil)

		reconciler := traffic_policy_aggregation.NewAggregationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			meshClient,
			aggregator,
			map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator{
				zephyr_core_types.MeshType_ISTIO: validator,
			},
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does not replace updated policies if they cannot merge with the existing state", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
		meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
		validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)

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
		mesh1 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-1",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster1,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}
		mesh2 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-2",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster2,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}
		meshServices := []*zephyr_discovery.MeshService{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms1",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh1.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: tp2LastValidState.Name,
							},
							TrafficPolicySpec: &tp2LastValidState.Spec,
						},
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: tps[0].Name,
							},
							TrafficPolicySpec: &tps[0].Spec,
						},
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms2",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh2.ObjectMeta),
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
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh1.ObjectMeta)).
			Return(mesh1, nil)
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh2.ObjectMeta)).
			Return(mesh2, nil)
		aggregator.EXPECT().
			GroupByMeshService(tps, gomock.Any()).
			DoAndReturn(func(
				trafficPolicies []*zephyr_networking.TrafficPolicy,
				meshServices []*zephyr_discovery.MeshService,
			) (result []*traffic_policy_aggregation.ServiceWithRelevantPolicies, err error) {
				// see https://github.com/solo-io/service-mesh-hub/issues/677 for why this is such a pain
				Expect(meshServices).To(HaveLen(2))
				var ret []*traffic_policy_aggregation.ServiceWithRelevantPolicies
				for _, service := range meshServices {
					var policies []*zephyr_networking.TrafficPolicy
					if service.ObjectMeta.Name == "ms1" {
						policies = trafficPolicies[0:2]
					} else {
						policies = trafficPolicies[2:4]
					}
					ret = append(ret, &traffic_policy_aggregation.ServiceWithRelevantPolicies{
						MeshService:     service,
						TrafficPolicies: policies,
					})
				}

				return ret, nil
			})
		aggregator.EXPECT().
			FindMergeConflict(&tps[1].Spec, []*types.TrafficPolicySpec{
				&tps[0].Spec,
			}, meshServices[0]).
			Return(conflictError)
		aggregator.EXPECT().
			FindMergeConflict(&tps[2].Spec, nil, meshServices[1]).
			Return(nil)
		aggregator.EXPECT().
			FindMergeConflict(&tps[3].Spec, []*types.TrafficPolicySpec{&tps[2].Spec}, meshServices[1]).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[0], mesh1, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{{
				Ref: &zephyr_core_types.ResourceRef{
					Name: tp2LastValidState.Name,
				},
				TrafficPolicySpec: &tp2LastValidState.Spec,
			}}).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[1], mesh2, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{{
				Ref: &zephyr_core_types.ResourceRef{
					Name: tps[2].Name,
				},
				TrafficPolicySpec: &tps[2].Spec,
			}}).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[1], mesh2, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: tps[2].Name,
					},
					TrafficPolicySpec: &tps[2].Spec,
				},
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: tps[3].Name,
					},
					TrafficPolicySpec: &tps[3].Spec,
				},
			}).
			Return(nil)

		// NOTE: No update issued for ms1

		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms2",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh2.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp3",
							},
							TrafficPolicySpec: &tps[2].Spec,
						},
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp4",
							},
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
			meshClient,
			aggregator,
			map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator{
				zephyr_core_types.MeshType_ISTIO: validator,
			},
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("resets traffic policy conflict errors if they no longer target any service", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
		meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
		validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)

		// going to associate tp1 and tp2 with service 1; and tp3 and tp4 with service 2
		tps := []*zephyr_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},

					// this should get cleared out later
					ConflictErrors: []*types.TrafficPolicyStatus_ConflictError{{
						ErrorMessage: "whoops error",
					}},
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
		mesh1 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-1",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster1,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}
		mesh2 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-2",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster2,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}

		meshServices := []*zephyr_discovery.MeshService{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms1",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh1.ObjectMeta),
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms2",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh2.ObjectMeta),
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
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh1.ObjectMeta)).
			Return(mesh1, nil)
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh2.ObjectMeta)).
			Return(mesh2, nil)
		aggregator.EXPECT().
			GroupByMeshService(tps, gomock.Any()).
			DoAndReturn(func(
				trafficPolicies []*zephyr_networking.TrafficPolicy,
				meshServices []*zephyr_discovery.MeshService,
			) (result []*traffic_policy_aggregation.ServiceWithRelevantPolicies, err error) {
				// see https://github.com/solo-io/service-mesh-hub/issues/677 for why this is such a pain
				Expect(meshServices).To(HaveLen(2))
				var ret []*traffic_policy_aggregation.ServiceWithRelevantPolicies
				for _, service := range meshServices {
					var policies []*zephyr_networking.TrafficPolicy
					if service.ObjectMeta.Name == "ms1" {
						policies = trafficPolicies[1:2] // intentionally excluding tp[0]
					} else {
						policies = trafficPolicies[2:4]
					}
					ret = append(ret, &traffic_policy_aggregation.ServiceWithRelevantPolicies{
						MeshService:     service,
						TrafficPolicies: policies,
					})
				}

				return ret, nil
			})
		aggregator.EXPECT().
			FindMergeConflict(&tps[1].Spec, nil, meshServices[0]).
			Return(nil)
		aggregator.EXPECT().
			FindMergeConflict(&tps[2].Spec, nil, meshServices[1]).
			Return(nil)
		aggregator.EXPECT().
			FindMergeConflict(&tps[3].Spec, []*types.TrafficPolicySpec{&tps[2].Spec}, meshServices[1]).
			Return(nil)
		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms1",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh1.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp2",
							},
							TrafficPolicySpec: &tps[1].Spec,
						},
					},
				},
			}).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[0], mesh1, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{{
				Ref: &zephyr_core_types.ResourceRef{
					Name: tps[1].Name,
				},
				TrafficPolicySpec: &tps[1].Spec,
			}}).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[1], mesh2, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: tps[2].Name,
					},
					TrafficPolicySpec: &tps[2].Spec,
				},
			}).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[1], mesh2, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: tps[2].Name,
					},
					TrafficPolicySpec: &tps[2].Spec,
				},
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: tps[3].Name,
					},
					TrafficPolicySpec: &tps[3].Spec,
				},
			}).
			Return(nil)
		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms2",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh2.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp3",
							},
							TrafficPolicySpec: &tps[2].Spec,
						},
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp4",
							},
							TrafficPolicySpec: &tps[3].Spec,
						},
					},
				},
			}).
			Return(nil)
		tp1Copy := *tps[0]
		tp1Copy.Status = types.TrafficPolicyStatus{
			ValidationStatus: tps[0].Status.ValidationStatus,
			ConflictErrors:   nil,
		}
		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &tp1Copy).
			Return(nil)

		reconciler := traffic_policy_aggregation.NewAggregationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			meshClient,
			aggregator,
			map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator{
				zephyr_core_types.MeshType_ISTIO: validator,
			},
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("does not include new policies if they cannot merge with the existing state", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
		meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
		validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)

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

		cluster1 := "cluster-1"
		cluster2 := "cluster-2"
		mesh1 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-1",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster1,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}
		mesh2 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-2",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster2,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}
		meshServices := []*zephyr_discovery.MeshService{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms1",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh1.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: tps[1].Name,
							},
							TrafficPolicySpec: &tps[1].Spec,
						},
					},
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms2",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh2.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: tps[2].Name,
							},
							TrafficPolicySpec: &tps[2].Spec,
						},
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: tps[3].Name,
							},
							TrafficPolicySpec: &tps[3].Spec,
						},
					},
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
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh1.ObjectMeta)).
			Return(mesh1, nil)
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh2.ObjectMeta)).
			Return(mesh2, nil)
		aggregator.EXPECT().
			GroupByMeshService(tps, gomock.Any()).
			DoAndReturn(func(
				trafficPolicies []*zephyr_networking.TrafficPolicy,
				meshServices []*zephyr_discovery.MeshService,
			) (result []*traffic_policy_aggregation.ServiceWithRelevantPolicies, err error) {
				// see https://github.com/solo-io/service-mesh-hub/issues/677 for why this is such a pain
				Expect(meshServices).To(HaveLen(2))
				var ret []*traffic_policy_aggregation.ServiceWithRelevantPolicies
				for _, service := range meshServices {
					var policies []*zephyr_networking.TrafficPolicy
					if service.ObjectMeta.Name == "ms1" {
						policies = trafficPolicies[0:2]
					} else {
						policies = trafficPolicies[2:4]
					}
					ret = append(ret, &traffic_policy_aggregation.ServiceWithRelevantPolicies{
						MeshService:     service,
						TrafficPolicies: policies,
					})
				}

				return ret, nil
			})

		aggregator.EXPECT().
			FindMergeConflict(&tps[0].Spec, []*types.TrafficPolicySpec{&tps[1].Spec}, meshServices[0]).
			Return(conflictError)

		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &zephyr_networking.TrafficPolicy{
				ObjectMeta: tps[0].ObjectMeta,
				Spec:       tps[0].Spec,
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
			meshClient,
			aggregator,
			map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator{
				zephyr_core_types.MeshType_ISTIO: validator,
			},
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("marks traffic policies as being in an error state if they cannot be translated", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		aggregator := mock_traffic_policy_aggregation.NewMockAggregator(ctrl)
		meshClient := mock_zephyr_discovery_clients.NewMockMeshClient(ctrl)
		validator := mock_mesh_translation.NewMockTranslationValidator(ctrl)

		// going to associate tp1 and tp2 with service 1; and tp3 and tp4 with service 2
		tps := []*zephyr_networking.TrafficPolicy{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{Name: "tp1"},
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
				},
				Spec: types.TrafficPolicySpec{
					Retries: &types.TrafficPolicySpec_RetryPolicy{
						Attempts: 1,
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
				Spec: types.TrafficPolicySpec{
					Retries: &types.TrafficPolicySpec_RetryPolicy{
						Attempts: 2,
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
				Spec: types.TrafficPolicySpec{
					Retries: &types.TrafficPolicySpec_RetryPolicy{
						Attempts: 3,
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
				Spec: types.TrafficPolicySpec{
					Retries: &types.TrafficPolicySpec_RetryPolicy{
						Attempts: 4,
					},
				},
			},
		}
		cluster1 := "cluster-1"
		cluster2 := "cluster-2"
		mesh1 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-1",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster1,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}
		mesh2 := &zephyr_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name: "mesh-2",
			},
			Spec: types2.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{
					Name: cluster2,
				},
				MeshType: &types2.MeshSpec_Istio{},
			},
		}

		meshServices := []*zephyr_discovery.MeshService{
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms1",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh1.ObjectMeta),
				},
			},
			{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms2",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh2.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: tps[2].Name,
							},
							TrafficPolicySpec: &tps[2].Spec,
						},
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: tps[3].Name,
							},
							TrafficPolicySpec: &types.TrafficPolicySpec{
								Retries: &types.TrafficPolicySpec_RetryPolicy{
									Attempts: 420,
								},
							},
						},
					},
				},
			},
		}
		translatorError := &types.TrafficPolicyStatus_TranslatorError{
			ErrorMessage: "translator-error",
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
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh1.ObjectMeta)).
			Return(mesh1, nil)
		meshClient.EXPECT().
			GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh2.ObjectMeta)).
			Return(mesh2, nil)
		aggregator.EXPECT().
			GroupByMeshService(tps, gomock.Any()).
			DoAndReturn(func(
				trafficPolicies []*zephyr_networking.TrafficPolicy,
				meshServices []*zephyr_discovery.MeshService,
			) (result []*traffic_policy_aggregation.ServiceWithRelevantPolicies, err error) {
				// see https://github.com/solo-io/service-mesh-hub/issues/677 for why this is such a pain
				Expect(meshServices).To(HaveLen(2))
				var ret []*traffic_policy_aggregation.ServiceWithRelevantPolicies
				for _, service := range meshServices {
					var policies []*zephyr_networking.TrafficPolicy
					if service.ObjectMeta.Name == "ms1" {
						policies = trafficPolicies[0:2]
					} else {
						policies = trafficPolicies[2:4]
					}
					ret = append(ret, &traffic_policy_aggregation.ServiceWithRelevantPolicies{
						MeshService:     service,
						TrafficPolicies: policies,
					})
				}

				return ret, nil
			})
		aggregator.EXPECT().
			FindMergeConflict(&tps[0].Spec, nil, meshServices[0]).
			Return(nil)
		aggregator.EXPECT().
			FindMergeConflict(&tps[1].Spec, nil, meshServices[0]).
			Return(nil)
		aggregator.EXPECT().
			FindMergeConflict(&tps[3].Spec, []*types.TrafficPolicySpec{&tps[2].Spec}, meshServices[1]).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[0], mesh1, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{{
				Ref: &zephyr_core_types.ResourceRef{
					Name: "tp1",
				},
				TrafficPolicySpec: &tps[0].Spec,
			}}).
			Return([]*mesh_translation.TranslationError{{
				Policy: &types2.MeshServiceStatus_ValidatedTrafficPolicy{
					Ref: &zephyr_core_types.ResourceRef{
						Name: tps[0].Name,
					},
					TrafficPolicySpec: &tps[0].Spec,
				},
				TranslatorErrors: []*types.TrafficPolicyStatus_TranslatorError{translatorError},
			}})
		validator.EXPECT().
			GetTranslationErrors(meshServices[0], mesh1, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: "tp2",
					},
					TrafficPolicySpec: &tps[1].Spec,
				},
			}).
			Return(nil)
		validator.EXPECT().
			GetTranslationErrors(meshServices[1], mesh2, []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: "tp3",
					},
					TrafficPolicySpec: &tps[2].Spec,
				},
				{
					Ref: &zephyr_core_types.ResourceRef{
						Name: "tp4",
					},
					TrafficPolicySpec: &tps[3].Spec,
				},
			}).
			Return([]*mesh_translation.TranslationError{{
				Policy: &types2.MeshServiceStatus_ValidatedTrafficPolicy{
					Ref: &zephyr_core_types.ResourceRef{
						Name: tps[3].Name,
					},
					TrafficPolicySpec: &tps[3].Spec,
				},
				TranslatorErrors: []*types.TrafficPolicyStatus_TranslatorError{translatorError},
			}})

		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms1",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh1.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp2",
							},
							TrafficPolicySpec: &tps[1].Spec,
						},
					},
				},
			}).
			Return(nil)
		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name: "ms2",
				},
				Spec: types2.MeshServiceSpec{
					Mesh: clients.ObjectMetaToResourceRef(mesh2.ObjectMeta),
				},
				Status: types2.MeshServiceStatus{
					ValidatedTrafficPolicies: []*types2.MeshServiceStatus_ValidatedTrafficPolicy{
						{
							Ref: &zephyr_core_types.ResourceRef{
								Name: "tp3",
							},
							TrafficPolicySpec: &tps[2].Spec,
						},
					},
				},
			}).
			Return(nil)
		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &zephyr_networking.TrafficPolicy{
				ObjectMeta: tps[0].ObjectMeta,
				Spec:       tps[0].Spec,
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
					TranslatorErrors: []*types.TrafficPolicyStatus_TranslatorError{translatorError},
				},
			}).
			Return(nil)
		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &zephyr_networking.TrafficPolicy{
				ObjectMeta: tps[3].ObjectMeta,
				Spec:       tps[3].Spec,
				Status: types.TrafficPolicyStatus{
					ValidationStatus: &zephyr_core_types.Status{
						State: zephyr_core_types.Status_ACCEPTED,
					},
					TranslatorErrors: []*types.TrafficPolicyStatus_TranslatorError{translatorError},
				},
			}).
			Return(nil)

		reconciler := traffic_policy_aggregation.NewAggregationReconciler(
			trafficPolicyClient,
			meshServiceClient,
			meshClient,
			aggregator,
			map[zephyr_core_types.MeshType]mesh_translation.TranslationValidator{
				zephyr_core_types.MeshType_ISTIO: validator,
			},
		)

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})
})
