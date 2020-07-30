package meshservice

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input/test"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	mock_authorizationpolicy "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/authorizationpolicy/mocks"
	mock_destinationrule "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/destinationrule/mocks"
	mock_virtualservice "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/istio/meshservice/virtualservice/mocks"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/security/v1beta1"
)

var _ = Describe("IstioMeshServiceTranslator", func() {
	var (
		ctrl                              *gomock.Controller
		mockDestinationRuleTranslator     *mock_destinationrule.MockTranslator
		mockVirtualServiceTranslator      *mock_virtualservice.MockTranslator
		mockAuthorizationPolicyTranslator *mock_authorizationpolicy.MockTranslator
		mockReporter                      *mock_reporting.MockReporter
		in                                input.Snapshot
		istioMeshServiceTranslator        Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDestinationRuleTranslator = mock_destinationrule.NewMockTranslator(ctrl)
		mockVirtualServiceTranslator = mock_virtualservice.NewMockTranslator(ctrl)
		mockAuthorizationPolicyTranslator = mock_authorizationpolicy.NewMockTranslator(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		in = test.NewInputSnapshotManualBuilder("").Build()
		istioMeshServiceTranslator = &translator{
			destinationRuleTranslator:     mockDestinationRuleTranslator,
			virtualServiceTranslator:      mockVirtualServiceTranslator,
			authorizationPolicyTranslator: mockAuthorizationPolicyTranslator,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
		meshService := &v1alpha2.MeshService{}

		expectedOutputs := Outputs{
			VirtualService:      &v1alpha3.VirtualService{},
			DestinationRule:     &v1alpha3.DestinationRule{},
			AuthorizationPolicy: &v1beta1.AuthorizationPolicy{},
		}

		mockDestinationRuleTranslator.
			EXPECT().
			Translate(in, meshService, mockReporter).
			Return(expectedOutputs.DestinationRule)
		mockVirtualServiceTranslator.
			EXPECT().
			Translate(in, meshService, mockReporter).
			Return(expectedOutputs.VirtualService)
		mockAuthorizationPolicyTranslator.
			EXPECT().
			Translate(in, meshService, mockReporter).
			Return(expectedOutputs.AuthorizationPolicy)

		outputs := istioMeshServiceTranslator.Translate(in, meshService, mockReporter)
		Expect(outputs).To(Equal(expectedOutputs))
	})
})
