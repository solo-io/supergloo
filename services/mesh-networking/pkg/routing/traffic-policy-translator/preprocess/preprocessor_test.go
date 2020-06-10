package preprocess_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	networking_selector "github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/common/kube/selection/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/errors"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
	mock_preprocess "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess/mocks"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Merger", func() {
	var (
		ctrl                 *gomock.Controller
		ctx                  context.Context
		mockMerger           *mock_preprocess.MockTrafficPolicyMerger
		mockValidator        *mock_preprocess.MockTrafficPolicyValidator
		mockResourceSelector *mock_selector.MockResourceSelector
		preprocessor         preprocess.TrafficPolicyPreprocessor
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockResourceSelector = mock_selector.NewMockResourceSelector(ctrl)
		mockMerger = mock_preprocess.NewMockTrafficPolicyMerger(ctrl)
		mockValidator = mock_preprocess.NewMockTrafficPolicyValidator(ctrl)
		preprocessor = preprocess.NewTrafficPolicyPreprocessor(
			mockResourceSelector,
			mockMerger,
			mockValidator,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should process TrafficPolicy", func() {
		selector := &smh_core_types.ServiceSelector{}
		namespace := "namespace"
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				DestinationSelector: selector,
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: namespace},
		}
		ms := []*smh_discovery.MeshService{}
		expectedMergedTPs := map[networking_selector.MeshServiceId][]*smh_networking.TrafficPolicy{}
		mockResourceSelector.
			EXPECT().
			GetAllMeshServicesByServiceSelector(ctx, selector).
			Return(ms, nil)
		mockMerger.
			EXPECT().
			MergeTrafficPoliciesForMeshServices(ctx, ms).
			Return(expectedMergedTPs, nil)
		mockValidator.EXPECT().Validate(ctx, tp).Return(nil)
		mergedTPs, err := preprocessor.PreprocessTrafficPolicy(ctx, tp)
		Expect(err).ToNot(HaveOccurred())
		Expect(mergedTPs).To(Equal(expectedMergedTPs))
	})

	It("should update triggering TrafficPolicy status to CONFLICT if conflict found during processing", func() {
		selector := &smh_core_types.ServiceSelector{}
		namespace := "namespace"
		tp := &smh_networking.TrafficPolicy{
			Spec: smh_networking_types.TrafficPolicySpec{
				DestinationSelector: selector,
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: namespace},
		}
		ms := []*smh_discovery.MeshService{}
		mockResourceSelector.
			EXPECT().
			GetAllMeshServicesByServiceSelector(ctx, selector).
			Return(ms, nil)
		mockMerger.
			EXPECT().
			MergeTrafficPoliciesForMeshServices(ctx, ms).
			Return(nil, errors.TrafficPolicyConflictError)
		mockValidator.EXPECT().Validate(ctx, tp).Return(nil)
		_, err := preprocessor.PreprocessTrafficPolicy(ctx, tp)
		Expect(err).To(testutils.HaveInErrorChain(errors.TrafficPolicyConflictError))
	})

	It("should process TrafficPolicies for MeshService", func() {
		ms := &smh_discovery.MeshService{}
		msList := []*smh_discovery.MeshService{ms}
		mergedTpsByMs := map[networking_selector.MeshServiceId][]*smh_networking.TrafficPolicy{}
		mockMerger.
			EXPECT().
			MergeTrafficPoliciesForMeshServices(ctx, msList).
			Return(mergedTpsByMs, nil)
		trafficPoliciesForMeshService, err := preprocessor.PreprocessTrafficPoliciesForMeshService(ctx, ms)
		Expect(err).ToNot(HaveOccurred())
		Expect(trafficPoliciesForMeshService).To(Equal(mergedTpsByMs))
	})
})
