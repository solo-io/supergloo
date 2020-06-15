package reconcile_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/reconciliation"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/framework"
	mock_traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/framework/mocks"
	. "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/reconcile"
	mock_traffic_policy_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot"
	mock_traffic_policy_snapshot "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/snapshot/mocks"
	mock_traffic_policy_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/validation/mocks"
	mock_smh_discovery_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_smh_networking_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.smh.solo.io/v1alpha1"
)

var _ = Describe("Reconcile", func() {
	var (
		ctx                  = context.TODO()
		ctrl                 *gomock.Controller
		trafficPolicyClient  *mock_smh_networking_clients.MockTrafficPolicyClient
		meshServiceClient    *mock_smh_discovery_clients.MockMeshServiceClient
		snapshotReconciler   *mock_traffic_policy_snapshot.MockTranslationSnapshotReconciler
		validationProcessor  *mock_traffic_policy_validation.MockValidationProcessor
		aggregationProcessor *mock_traffic_policy_aggregation.MockAggregationProcessor
		translationProcessor *mock_traffic_policy_translation.MockTranslationProcessor
		reconciler           reconciliation.Reconciler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		trafficPolicyClient = mock_smh_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient = mock_smh_discovery_clients.NewMockMeshServiceClient(ctrl)

		snapshotReconciler = mock_traffic_policy_snapshot.NewMockTranslationSnapshotReconciler(ctrl)
		validationProcessor = mock_traffic_policy_validation.NewMockValidationProcessor(ctrl)
		aggregationProcessor = mock_traffic_policy_aggregation.NewMockAggregationProcessor(ctrl)
		translationProcessor = mock_traffic_policy_translation.NewMockTranslationProcessor(ctrl)
		reconciler = NewReconciler(
			trafficPolicyClient,
			meshServiceClient,
			snapshotReconciler,
			validationProcessor,
			aggregationProcessor,
			translationProcessor,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("can set a new validation status", func() {
		invalidTrafficPolicy := &smh_networking.TrafficPolicy{}
		failedValidationStatus := &smh_core_types.Status{
			State: smh_core_types.Status_INVALID,
		}
		updatedTrafficPolicy := &smh_networking.TrafficPolicy{
			Status: types.TrafficPolicyStatus{
				ValidationStatus: failedValidationStatus,
			},
		}

		trafficPolicyClient.EXPECT().
			ListTrafficPolicy(ctx).
			Return(&smh_networking.TrafficPolicyList{
				Items: []smh_networking.TrafficPolicy{*invalidTrafficPolicy},
			}, nil)

		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&v1alpha1.MeshServiceList{}, nil)

		validationProcessor.EXPECT().Process(ctx, gomock.Any(), gomock.Any()).
			Return([]*smh_networking.TrafficPolicy{updatedTrafficPolicy})
		aggregationProcessor.EXPECT().Process(ctx, gomock.Any())
		snap := snapshot.ClusterNameToSnapshot{}
		translationProcessor.EXPECT().Process(ctx, gomock.Any()).Return(snap, nil)
		snapshotReconciler.EXPECT().ReconcileAllSnapshots(ctx, snap)

		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &smh_networking.TrafficPolicy{
				Status: types.TrafficPolicyStatus{
					ValidationStatus: failedValidationStatus,
				},
			})

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can set a new aggregated status", func() {
		invalidTrafficPolicy1 := &smh_networking.TrafficPolicy{
			ObjectMeta: v1.ObjectMeta{Name: "tp1"},
		}
		invalidTrafficPolicy2 := &smh_networking.TrafficPolicy{
			ObjectMeta: v1.ObjectMeta{Name: "tp2"},
		}
		failedValidationStatus := &smh_core_types.Status{
			State: smh_core_types.Status_INVALID,
		}
		updatedTrafficPolicy := func(t *smh_networking.TrafficPolicy) *smh_networking.TrafficPolicy {
			t.Status = types.TrafficPolicyStatus{
				ValidationStatus: failedValidationStatus,
			}
			return t
		}
		meshService := &smh_discovery.MeshService{
			Status: smh_discovery_types.MeshServiceStatus{
				ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
					Ref: &smh_core_types.ResourceRef{Name: "validated-1"},
				}},
			},
		}
		trafficPolicyClient.EXPECT().
			ListTrafficPolicy(ctx).
			Return(&smh_networking.TrafficPolicyList{
				Items: []smh_networking.TrafficPolicy{*invalidTrafficPolicy1, *invalidTrafficPolicy2},
			}, nil)

		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&v1alpha1.MeshServiceList{Items: []smh_discovery.MeshService{*meshService}}, nil)

		validationProcessor.EXPECT().Process(ctx, gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, allTrafficPolicies []*smh_networking.TrafficPolicy, meshServices []*smh_discovery.MeshService) []*smh_networking.TrafficPolicy {
				updatedTrafficPolicy(allTrafficPolicies[0])
				return []*smh_networking.TrafficPolicy{updatedTrafficPolicy(allTrafficPolicies[0])}
			})
		aggregationProcessor.EXPECT().Process(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, allTrafficPolicies []*smh_networking.TrafficPolicy) (*traffic_policy_aggregation.ProcessedObjects, error) {
				updatedTrafficPolicy(allTrafficPolicies[0])
				updatedTrafficPolicy(allTrafficPolicies[1])
				return &traffic_policy_aggregation.ProcessedObjects{
					TrafficPolicies: allTrafficPolicies,
					MeshServices:    []*smh_discovery.MeshService{meshService},
				}, nil
			})

		snap := snapshot.ClusterNameToSnapshot{}
		translationProcessor.EXPECT().Process(ctx, gomock.Any()).Return(snap, nil)
		snapshotReconciler.EXPECT().ReconcileAllSnapshots(ctx, snap)

		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, updatedTrafficPolicy(invalidTrafficPolicy1))
		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, updatedTrafficPolicy(invalidTrafficPolicy2))
		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &smh_discovery.MeshService{
				Status: smh_discovery_types.MeshServiceStatus{
					ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
						Ref: &smh_core_types.ResourceRef{Name: "validated-1"},
					}},
				},
			})

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("errors when translation fails, but still updates other statues", func() {
		invalidTrafficPolicy := &smh_networking.TrafficPolicy{}
		failedValidationStatus := &smh_core_types.Status{
			State: smh_core_types.Status_INVALID,
		}
		updatedTrafficPolicy := &smh_networking.TrafficPolicy{
			Status: types.TrafficPolicyStatus{
				ValidationStatus: failedValidationStatus,
			},
		}

		meshService := &smh_discovery.MeshService{
			Status: smh_discovery_types.MeshServiceStatus{
				ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
					Ref: &smh_core_types.ResourceRef{Name: "validated-1"},
				}},
			},
		}
		trafficPolicyClient.EXPECT().
			ListTrafficPolicy(ctx).
			Return(&smh_networking.TrafficPolicyList{
				Items: []smh_networking.TrafficPolicy{*invalidTrafficPolicy},
			}, nil)

		meshServiceClient.EXPECT().
			ListMeshService(ctx).
			Return(&v1alpha1.MeshServiceList{Items: []smh_discovery.MeshService{*meshService}}, nil)

		validationProcessor.EXPECT().Process(ctx, gomock.Any(), gomock.Any()).
			Return([]*smh_networking.TrafficPolicy{updatedTrafficPolicy})
		aggregationProcessor.EXPECT().Process(ctx, gomock.Any()).Return(&traffic_policy_aggregation.ProcessedObjects{
			MeshServices: []*smh_discovery.MeshService{meshService},
		}, nil)

		translationProcessor.EXPECT().Process(ctx, gomock.Any()).Return(nil, errors.New("translation error"))

		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &smh_networking.TrafficPolicy{
				Status: types.TrafficPolicyStatus{
					ValidationStatus: failedValidationStatus,
				},
			})
		meshServiceClient.EXPECT().
			UpdateMeshServiceStatus(ctx, &smh_discovery.MeshService{
				Status: smh_discovery_types.MeshServiceStatus{
					ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
						Ref: &smh_core_types.ResourceRef{Name: "validated-1"},
					}},
				},
			})
		err := reconciler.Reconcile(ctx)
		Expect(err).To(MatchError(ContainSubstring("translation error")))
	})
})
