package osm

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	mock_output "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/smi/mocks"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm/internal/mocks"
	mock_mesh "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm/mesh/mocks"
	mock_traffictarget "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm/traffictarget/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("SmiNetworkingTranslator", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		mockReporter                *mock_reporting.MockReporter
		mockOutputs                 *mock_output.MockBuilder
		mockDependencyFactory       *MockDependencyFactory
		mockMeshTranslator          *mock_mesh.MockTranslator
		mockTrafficTargetTranslator *mock_traffictarget.MockTranslator
		translator                  *osmTranslator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockDependencyFactory = NewMockDependencyFactory(ctrl)
		mockOutputs = mock_output.NewMockBuilder(ctrl)
		mockMeshTranslator = mock_mesh.NewMockTranslator(ctrl)
		mockTrafficTargetTranslator = mock_traffictarget.NewMockTranslator(ctrl)
		translator = &osmTranslator{dependencies: mockDependencyFactory}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate all meshes and traffictargets", func() {
		in := input.NewInputSnapshotManualBuilder("").
			AddMeshes([]*discoveryv1alpha2.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       discoveryv1alpha2.MeshSpec{},
					Status:     discoveryv1alpha2.MeshStatus{},
				},
			}).
			AddTrafficTargets([]*discoveryv1alpha2.TrafficTarget{
				{
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       discoveryv1alpha2.TrafficTargetSpec{},
					Status:     discoveryv1alpha2.TrafficTargetStatus{},
				},
			}).
			Build()

		mockDependencyFactory.
			EXPECT().
			MakeMeshTranslator().
			Return(mockMeshTranslator)

		mockDependencyFactory.
			EXPECT().
			MakeTrafficTargetTranslator().
			Return(mockTrafficTargetTranslator)

		for i := range in.Meshes().List() {
			mockMeshTranslator.
				EXPECT().
				Translate(gomock.Any(), in, in.Meshes().List()[i], mockOutputs, mockReporter)
		}

		for i := range in.TrafficTargets().List() {
			mockTrafficTargetTranslator.
				EXPECT().
				Translate(gomock.Any(), in, in.TrafficTargets().List()[i], mockOutputs, mockReporter)
		}

		translator.Translate(ctx, in, mockOutputs, mockReporter)
	})
})
