package reconcile_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/reconcile"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation"
	mock_failover_service_translation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation/mocks"
	mock_failover_service_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/validation/mocks"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Processor", func() {
	var (
		ctrl           *gomock.Controller
		ctx            context.Context
		mockValidator  *mock_failover_service_validation.MockFailoverServiceValidator
		mockTranslator *mock_failover_service_translation.MockFailoverServiceTranslator
		processor      reconcile.FailoverServiceProcessor
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockValidator = mock_failover_service_validation.NewMockFailoverServiceValidator(ctrl)
		mockTranslator = mock_failover_service_translation.NewMockFailoverServiceTranslator(ctrl)
		processor = reconcile.NewFailoverServiceProcessor(
			mockValidator, []translation.FailoverServiceTranslator{
				mockTranslator,
			})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var buildInputSnapshot = func() failover.InputSnapshot {
		return failover.InputSnapshot{
			FailoverServices: []*v1alpha1.FailoverService{
				// Should be skipped because ObservedGeneration != Generation.
				{
					ObjectMeta: v1.ObjectMeta{
						Generation: 2,
					},
					Status: types.FailoverServiceStatus{
						ObservedGeneration: 1,
					},
				},
				// Should be skipped because not validated.
				{
					Status: types.FailoverServiceStatus{
						ValidationStatus: &smh_core_types.Status{
							State: smh_core_types.Status_INVALID,
						},
					},
				},
				{
					ObjectMeta: v1.ObjectMeta{
						Generation: 1,
					},
					Spec: types.FailoverServiceSpec{
						Services: []*smh_core_types.ResourceRef{
							{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster1",
							},
							{
								Name:      "service2",
								Namespace: "namespace2",
								Cluster:   "cluster2",
							},
						},
					},
					Status: types.FailoverServiceStatus{
						ObservedGeneration: 1,
						ValidationStatus: &smh_core_types.Status{
							State: smh_core_types.Status_ACCEPTED,
						},
					},
				},
			},
			MeshServices: []*smh_discovery.MeshService{
				{
					Spec: types2.MeshServiceSpec{
						KubeService: &types2.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster1",
							},
						},
					},
				},
				{
					Spec: types2.MeshServiceSpec{
						KubeService: &types2.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "service2",
								Namespace: "namespace2",
								Cluster:   "cluster2",
							},
						},
					},
				},
			},
		}
	}

	It("should process input snapshot", func() {
		inputSnapshot := buildInputSnapshot()
		mockValidator.EXPECT().Validate(inputSnapshot)
		expectedServiceEntries := []*v1alpha3.ServiceEntry{{ObjectMeta: v1.ObjectMeta{Name: "se1"}}}
		expectedEnvoyFilters := []*v1alpha3.EnvoyFilter{{ObjectMeta: v1.ObjectMeta{Name: "dr1"}}}
		mockTranslator.
			EXPECT().
			Translate(ctx, inputSnapshot.FailoverServices[2], inputSnapshot.MeshServices).
			Return(failover.MeshOutputs{
				ServiceEntries: expectedServiceEntries,
				EnvoyFilters:   expectedEnvoyFilters,
			}, nil)
		expectedOutputFailoverService := *inputSnapshot.FailoverServices[2]
		expectedOutputFailoverService.Status.TranslationStatus = &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
		outputSnapshot := processor.Process(ctx, inputSnapshot)
		Expect(outputSnapshot.MeshOutputs.ServiceEntries).To(Equal(expectedServiceEntries))
		Expect(outputSnapshot.MeshOutputs.EnvoyFilters).To(Equal(expectedEnvoyFilters))
		Expect(outputSnapshot.FailoverServices).To(Equal([]*v1alpha1.FailoverService{&expectedOutputFailoverService}))
	})

	It("should process input snapshot with errors", func() {
		translatorError := &types.FailoverServiceStatus_TranslatorError{
			TranslatorId: "mock-translator",
			ErrorMessage: "test-error",
		}
		inputSnapshot := buildInputSnapshot()
		mockValidator.EXPECT().Validate(inputSnapshot)
		mockTranslator.
			EXPECT().
			Translate(ctx, inputSnapshot.FailoverServices[2], inputSnapshot.MeshServices).
			Return(failover.MeshOutputs{}, translatorError)
		expectedOutputFailoverService := *inputSnapshot.FailoverServices[2]
		expectedOutputFailoverService.Status.TranslationStatus = &smh_core_types.Status{
			State: smh_core_types.Status_PROCESSING_ERROR,
		}
		expectedOutputFailoverService.Status.TranslatorErrors = []*types.FailoverServiceStatus_TranslatorError{translatorError}
		outputSnapshot := processor.Process(ctx, inputSnapshot)
		Expect(outputSnapshot.MeshOutputs.ServiceEntries).To(BeEmpty())
		Expect(outputSnapshot.MeshOutputs.EnvoyFilters).To(BeEmpty())
		Expect(outputSnapshot.FailoverServices).To(Equal([]*v1alpha1.FailoverService{&expectedOutputFailoverService}))
	})
})
