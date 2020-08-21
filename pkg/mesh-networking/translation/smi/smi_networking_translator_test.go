package smi_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/istio/input"
	mock_output "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/mocks"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/internal/mocks"
	mock_meshservice "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/meshservice/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/utils/metautils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("SmiNetworkingTranslator", func() {
	var (
		ctrl                      *gomock.Controller
		ctx                       context.Context
		mockReporter              *mock_reporting.MockReporter
		mockOutputs               *mock_output.MockBuilder
		mockMeshServiceTranslator *mock_meshservice.MockTranslator
		mockDependencyFactory     *MockDependencyFactory
		translator                Translator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockMeshServiceTranslator = mock_meshservice.NewMockTranslator(ctrl)
		mockDependencyFactory = NewMockDependencyFactory(ctrl)
		mockOutputs = mock_output.NewMockBuilder(ctrl)
		translator = NewSmiTranslator(mockDependencyFactory)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate", func() {
		in := input.NewInputSnapshotManualBuilder("").
			AddMeshServices([]*discoveryv1alpha2.MeshService{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "mesh-service-1",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "mesh-service-2",
						Labels: metautils.TranslatedObjectLabels(),
					},
				},
			}).Build()

		mockDependencyFactory.
			EXPECT().
			MakeMeshServiceTranslator().
			Return(mockMeshServiceTranslator)
		for i := range in.MeshServices().List() {
			mockMeshServiceTranslator.
				EXPECT().
				Translate(gomock.Any(), in, in.MeshServices().List()[i], mockOutputs, mockReporter)
		}

		translator.Translate(ctx, in, mockOutputs, mockReporter)
	})
})
