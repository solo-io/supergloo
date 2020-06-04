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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
})
