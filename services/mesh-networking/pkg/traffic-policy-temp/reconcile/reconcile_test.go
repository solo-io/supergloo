package reconcile_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/common/reconciliation"
	mock_traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation/framework/mocks"
	. "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/reconcile"
	mock_traffic_policy_translation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/translation/framework/mocks"
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
		aggregationProcessor *mock_traffic_policy_aggregation.MockAggregateProcessor
		translationProcessor *mock_traffic_policy_translation.MockTranslationProcessor
		reconciler           reconciliation.Reconciler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		trafficPolicyClient = mock_smh_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient = mock_smh_discovery_clients.NewMockMeshServiceClient(ctrl)

		snapshotReconciler = mock_traffic_policy_snapshot.NewMockTranslationSnapshotReconciler(ctrl)
		validationProcessor = mock_traffic_policy_validation.NewMockValidationProcessor(ctrl)
		aggregationProcessor = mock_traffic_policy_aggregation.NewMockAggregateProcessor(ctrl)
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
		translationProcessor.EXPECT().Process(ctx)
		snapshotReconciler.EXPECT().ReconcileAllSnapshots(ctx, gomock.Any())

		trafficPolicyClient.EXPECT().
			UpdateTrafficPolicyStatus(ctx, &smh_networking.TrafficPolicy{
				Status: types.TrafficPolicyStatus{
					ValidationStatus: failedValidationStatus,
				},
			})

		err := reconciler.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

})
