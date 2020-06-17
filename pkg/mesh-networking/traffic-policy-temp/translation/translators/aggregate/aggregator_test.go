package mesh_translation_test

import (
	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	mock_mesh_translation "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/translators/mocks"

	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/traffic-policy-temp/translation/translators/aggregate"
)

var _ = Describe("Aggregator", func() {

	var (
		ctrl                  *gomock.Controller
		translator            *mock_mesh_translation.MockIstioTranslator
		meshTranslatorFactory *MeshTranslatorFactory
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		translator = mock_mesh_translation.NewMockIstioTranslator(ctrl)
		meshTranslatorFactory = NewMeshTranslatorFactory(translator)
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	It("should work for istio 1.5", func() {
		Expect(meshTranslatorFactory.MeshTypeToAccumulator(smh_core_types.MeshType_ISTIO1_5)).To(BeIdenticalTo(translator))
		Expect(meshTranslatorFactory.MeshTypeToTranslationValidator(smh_core_types.MeshType_ISTIO1_5)).To(BeIdenticalTo(translator))
	})
	It("should work for istio 1.6", func() {
		Expect(meshTranslatorFactory.MeshTypeToAccumulator(smh_core_types.MeshType_ISTIO1_6)).To(BeIdenticalTo(translator))
		Expect(meshTranslatorFactory.MeshTypeToTranslationValidator(smh_core_types.MeshType_ISTIO1_6)).To(BeIdenticalTo(translator))
	})

})
