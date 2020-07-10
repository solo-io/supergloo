package reconcile_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1alpha3sets "github.com/solo-io/external-apis/pkg/api/istio/networking.istio.io/v1alpha3/sets"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	types2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/reconcile"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation"
	mock_failover_service_translation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/translation/mocks"
	mock_failover_service_validation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/validation/mocks"
	v12 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
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
			FailoverServices: v1alpha1sets.NewFailoverServiceSet(
				// Should be skipped because ObservedGeneration != Generation.
				&v1alpha1.FailoverService{
					ObjectMeta: v1.ObjectMeta{
						Name:       "fs1",
						Generation: 2,
					},
					Status: types.FailoverServiceStatus{
						ObservedGeneration: 1,
					},
				},
				// Should be skipped because not validated.
				&v1alpha1.FailoverService{
					ObjectMeta: v1.ObjectMeta{
						Name: "fs2",
					},
					Status: types.FailoverServiceStatus{
						ValidationStatus: &smh_core_types.Status{
							State: smh_core_types.Status_INVALID,
						},
					},
				},
				&v1alpha1.FailoverService{
					ObjectMeta: v1.ObjectMeta{
						Name:       "fs3",
						Generation: 1,
					},
					Spec: types.FailoverServiceSpec{
						Hostname: "service1.namespace1.cluster1",
						Port: &types.FailoverServiceSpec_Port{
							Port:     9080,
							Protocol: "http",
						},
						Meshes: []*v12.ObjectRef{
							{
								Name:      "mesh1",
								Namespace: "namespace1",
							},
						},
						FailoverServices: []*v12.ClusterObjectRef{
							{
								Name:        "service2",
								Namespace:   "namespace2",
								ClusterName: "cluster2",
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
			),
			MeshServices: v1alpha1sets2.NewMeshServiceSet(
				&smh_discovery.MeshService{
					ObjectMeta: v1.ObjectMeta{
						Name:        "service2",
						Namespace:   "namespace2",
						ClusterName: "cluster2",
					},
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
			),
			Meshes: v1alpha1sets2.NewMeshSet(
				&smh_discovery.Mesh{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mesh1",
						Namespace: "namespace1",
					},
					Spec: types2.MeshSpec{
						Cluster: &smh_core_types.ResourceRef{Name: "cluster1"},
					},
				},
			),
		}
	}

	It("should process input snapshot", func() {
		inputSnapshot := buildInputSnapshot()
		mockValidator.EXPECT().Validate(inputSnapshot)
		expectedServiceEntries := v1alpha3sets.NewServiceEntrySet(&v1alpha3.ServiceEntry{ObjectMeta: v1.ObjectMeta{Name: "se1"}})
		expectedEnvoyFilters := v1alpha3sets.NewEnvoyFilterSet(&v1alpha3.EnvoyFilter{ObjectMeta: v1.ObjectMeta{Name: "dr1"}})
		mockTranslator.
			EXPECT().
			Translate(ctx, inputSnapshot.FailoverServices.List()[2], inputSnapshot.MeshServices.List(), inputSnapshot.Meshes).
			Return(failover.MeshOutputs{
				ServiceEntries: expectedServiceEntries,
				EnvoyFilters:   expectedEnvoyFilters,
			}, nil)
		expectedOutputFailoverService := *inputSnapshot.FailoverServices.List()[2]
		expectedOutputFailoverService.Status.TranslationStatus = &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
		outputSnapshot := processor.Process(ctx, inputSnapshot)
		Expect(outputSnapshot.MeshOutputs.ServiceEntries.List()).To(Equal(expectedServiceEntries.List()))
		Expect(outputSnapshot.MeshOutputs.EnvoyFilters.List()).To(Equal(expectedEnvoyFilters.List()))
		Expect(outputSnapshot.FailoverServices.List()).To(Equal([]*v1alpha1.FailoverService{
			inputSnapshot.FailoverServices.List()[0],
			inputSnapshot.FailoverServices.List()[1],
			&expectedOutputFailoverService,
		}))
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
			Translate(ctx, inputSnapshot.FailoverServices.List()[2], inputSnapshot.MeshServices.List(), inputSnapshot.Meshes).
			Return(failover.MeshOutputs{}, translatorError)
		expectedOutputFailoverService := *inputSnapshot.FailoverServices.List()[2]
		expectedOutputFailoverService.Status.TranslationStatus = &smh_core_types.Status{
			State: smh_core_types.Status_PROCESSING_ERROR,
		}
		expectedOutputFailoverService.Status.TranslatorErrors = []*types.FailoverServiceStatus_TranslatorError{translatorError}
		outputSnapshot := processor.Process(ctx, inputSnapshot)
		Expect(outputSnapshot.MeshOutputs.ServiceEntries.List()).To(BeEmpty())
		Expect(outputSnapshot.MeshOutputs.EnvoyFilters.List()).To(BeEmpty())
		Expect(outputSnapshot.FailoverServices.List()).To(Equal([]*v1alpha1.FailoverService{
			inputSnapshot.FailoverServices.List()[0],
			inputSnapshot.FailoverServices.List()[1],
			&expectedOutputFailoverService,
		}))
	})
})
