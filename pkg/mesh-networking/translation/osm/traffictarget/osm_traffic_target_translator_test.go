package traffictarget_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	mock_output "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/smi/mocks"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/osm/traffictarget"
	mock_traffictarget "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget/mocks"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("SmiTrafficTargetTranslator", func() {
	var (
		ctx                        context.Context
		ctrl                       *gomock.Controller
		mockOutputs                *mock_output.MockBuilder
		mockReporter               *mock_reporting.MockReporter
		mockSmiTranslator          *mock_traffictarget.MockTranslator
		osmTrafficTargetTranslator Translator
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockOutputs = mock_output.NewMockBuilder(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockSmiTranslator = mock_traffictarget.NewMockTranslator(ctrl)
		osmTrafficTargetTranslator = NewTranslator(mockSmiTranslator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should not translate when not an osm traffic target", func() {
		in := input.NewInputSnapshotManualBuilder("").Build()
		trafficTarget := &v1alpha2.TrafficTarget{}

		osmTrafficTargetTranslator.Translate(ctx, in, trafficTarget, mockOutputs, mockReporter)
	})

	It("should translate when an osm traffic target", func() {
		trafficTarget := &v1alpha2.TrafficTarget{
			Spec: v1alpha2.TrafficTargetSpec{
				Mesh: &v1.ObjectRef{
					Name:      "hello",
					Namespace: "world",
				},
			},
		}
		in := input.NewInputSnapshotManualBuilder("").
			AddMeshes([]*v1alpha2.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      trafficTarget.Spec.GetMesh().GetName(),
						Namespace: trafficTarget.Spec.GetMesh().GetNamespace(),
					},
					Spec: v1alpha2.MeshSpec{
						MeshType: &v1alpha2.MeshSpec_Osm{
							Osm: &v1alpha2.MeshSpec_OSM{},
						},
					},
				},
			}).
			Build()

		mockSmiTranslator.
			EXPECT().
			Translate(gomock.AssignableToTypeOf(ctx), in, trafficTarget, mockOutputs, mockReporter)

		osmTrafficTargetTranslator.Translate(ctx, in, trafficTarget, mockOutputs, mockReporter)
	})
})
