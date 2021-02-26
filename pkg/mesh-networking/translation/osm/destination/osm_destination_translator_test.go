package destination_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	v1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	mock_output "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/output/smi/mocks"
	mock_reporting "github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting/mocks"
	. "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/osm/destination"
	mock_destination "github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/smi/destination/mocks"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("SmiDestinationTranslator", func() {
	var (
		ctx                      context.Context
		ctrl                     *gomock.Controller
		mockOutputs              *mock_output.MockBuilder
		mockReporter             *mock_reporting.MockReporter
		mockSmiTranslator        *mock_destination.MockTranslator
		osmDestinationTranslator Translator
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockOutputs = mock_output.NewMockBuilder(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockSmiTranslator = mock_destination.NewMockTranslator(ctrl)
		osmDestinationTranslator = NewTranslator(mockSmiTranslator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should not translate when not an osm destination", func() {
		in := input.NewInputLocalSnapshotManualBuilder("").Build()
		destination := &v1.Destination{}

		osmDestinationTranslator.Translate(ctx, in, destination, mockOutputs, mockReporter)
	})

	It("should translate when an osm destination", func() {
		destination := &v1.Destination{
			Spec: v1.DestinationSpec{
				Mesh: &skv2corev1.ObjectRef{
					Name:      "hello",
					Namespace: "world",
				},
			},
		}
		in := input.NewInputLocalSnapshotManualBuilder("").
			AddMeshes([]*v1.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      destination.Spec.GetMesh().GetName(),
						Namespace: destination.Spec.GetMesh().GetNamespace(),
					},
					Spec: v1.MeshSpec{
						Type: &v1.MeshSpec_Osm{
							Osm: &v1.MeshSpec_OSM{},
						},
					},
				},
			}).
			Build()

		mockSmiTranslator.
			EXPECT().
			Translate(gomock.AssignableToTypeOf(ctx), in, destination, mockOutputs, mockReporter)

		osmDestinationTranslator.Translate(ctx, in, destination, mockOutputs, mockReporter)
	})
})
