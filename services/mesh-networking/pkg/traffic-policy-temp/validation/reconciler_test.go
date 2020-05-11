package traffic_policy_validation_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	v1alpha12 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	traffic_policy_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/validation"
	mock_traffic_policy_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/validation/mocks"
	mock_zephyr_discovery_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_zephyr_networking_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
)

var _ = Describe("Validation Reconciler", func() {
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

	It("can set a new validation status", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		validator := mock_traffic_policy_validation.NewMockValidator(ctrl)

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
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		validator := mock_traffic_policy_validation.NewMockValidator(ctrl)

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

	It("zeroes out any existing conflict status", func() {
		trafficPolicyClient := mock_zephyr_networking_clients.NewMockTrafficPolicyClient(ctrl)
		meshServiceClient := mock_zephyr_discovery_clients.NewMockMeshServiceClient(ctrl)
		validator := mock_traffic_policy_validation.NewMockValidator(ctrl)

		invalidTrafficPolicy := &v1alpha12.TrafficPolicy{
			Status: types.TrafficPolicyStatus{
				ValidationStatus: &zephyr_core_types.Status{
					State: zephyr_core_types.Status_ACCEPTED,
				},
				ConflictErrors: []*types.TrafficPolicyStatus_ConflictError{{
					ErrorMessage: "whoops there's a conflict",
				}},
			},
		}
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
})
