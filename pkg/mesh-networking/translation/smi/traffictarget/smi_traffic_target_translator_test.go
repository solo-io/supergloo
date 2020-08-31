package traffictarget_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	smiaccessv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha2"
	smispecsv1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha3"
	smisplitv1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	mock_output "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/output/smi/mocks"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget"
	mock_access "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget/access/mocks"
	mock_split "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/traffictarget/split/mocks"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
)

var _ = Describe("SmiTrafficTargetTranslator", func() {
	var (
		ctx                        context.Context
		ctrl                       *gomock.Controller
		mockOutputs                *mock_output.MockBuilder
		mockReporter               *mock_reporting.MockReporter
		mockSplitTranslator        *mock_split.MockTranslator
		mockAccessTranslator       *mock_access.MockTranslator
		smiTrafficTargetTranslator Translator
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.Background(), GinkgoT())
		mockOutputs = mock_output.NewMockBuilder(ctrl)
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		mockSplitTranslator = mock_split.NewMockTranslator(ctrl)
		mockAccessTranslator = mock_access.NewMockTranslator(ctrl)
		smiTrafficTargetTranslator = NewTranslator(mockSplitTranslator, mockAccessTranslator)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should translate when an smi traffic target", func() {
		trafficTarget := &v1alpha2.TrafficTarget{
			Spec: v1alpha2.TrafficTargetSpec{
				Mesh: &v1.ObjectRef{
					Name:      "hello",
					Namespace: "world",
				},
			},
		}
		in := input.NewInputSnapshotManualBuilder("").Build()

		ts := &smisplitv1alpha2.TrafficSplit{}

		mockSplitTranslator.
			EXPECT().
			Translate(gomock.AssignableToTypeOf(ctx), in, trafficTarget, mockReporter).
			Return(ts)

		mockOutputs.
			EXPECT().
			AddTrafficSplits(ts)

		tt := &smiaccessv1alpha2.TrafficTarget{}
		hrg := &smispecsv1alpha3.HTTPRouteGroup{}
		mockAccessTranslator.
			EXPECT().
			Translate(gomock.AssignableToTypeOf(ctx), in, trafficTarget, mockReporter).
			Return([]*smiaccessv1alpha2.TrafficTarget{tt}, []*smispecsv1alpha3.HTTPRouteGroup{hrg})

		mockOutputs.
			EXPECT().
			AddTrafficTargets(tt)

		mockOutputs.
			EXPECT().
			AddHTTPRouteGroups(hrg)

		smiTrafficTargetTranslator.Translate(ctx, in, trafficTarget, mockOutputs, mockReporter)
	})
})
