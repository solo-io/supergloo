package preprocess_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	networking_selector "github.com/solo-io/mesh-projects/pkg/selector"
	mock_selector "github.com/solo-io/mesh-projects/pkg/selector/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/errors"
	"github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess"
	mock_preprocess "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/routing/traffic-policy-translator/preprocess/mocks"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		selector := &core_types.ServiceSelector{}
		namespace := "namespace"
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: types.TrafficPolicySpec{
				DestinationSelector: selector,
			},
			ObjectMeta: v1.ObjectMeta{Namespace: namespace},
		}
		ms := []*discovery_v1alpha1.MeshService{}
		expectedMergedTPs := map[networking_selector.MeshServiceId][]*networking_v1alpha1.TrafficPolicy{}
		mockResourceSelector.
			EXPECT().
			GetMeshServicesByServiceSelector(ctx, selector).
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
		selector := &core_types.ServiceSelector{}
		namespace := "namespace"
		tp := &networking_v1alpha1.TrafficPolicy{
			Spec: types.TrafficPolicySpec{
				DestinationSelector: selector,
			},
			ObjectMeta: v1.ObjectMeta{Namespace: namespace},
		}
		ms := []*discovery_v1alpha1.MeshService{}
		mockResourceSelector.
			EXPECT().
			GetMeshServicesByServiceSelector(ctx, selector).
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
		ms := &discovery_v1alpha1.MeshService{}
		msList := []*discovery_v1alpha1.MeshService{ms}
		mergedTpsByMs := map[networking_selector.MeshServiceId][]*networking_v1alpha1.TrafficPolicy{}
		mockMerger.
			EXPECT().
			MergeTrafficPoliciesForMeshServices(ctx, msList).
			Return(mergedTpsByMs, nil)
		trafficPoliciesForMeshService, err := preprocessor.PreprocessTrafficPoliciesForMeshService(ctx, ms)
		Expect(err).ToNot(HaveOccurred())
		Expect(trafficPoliciesForMeshService).To(Equal(mergedTpsByMs))
	})
})
