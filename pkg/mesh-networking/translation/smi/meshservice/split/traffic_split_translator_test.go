package split_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/input"
	mock_reporting "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/reporting/mocks"
	. "github.com/solo-io/service-mesh-hub/pkg/mesh-networking/translation/smi/meshservice/split"
)

var _ = Describe("TrafficSplitTranslator", func() {
	var (
		ctrl                   *gomock.Controller
		mockReporter           *mock_reporting.MockReporter
		trafficSplitTranslator Translator
		in                     input.Snapshot
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockReporter = mock_reporting.NewMockReporter(ctrl)
		trafficSplitTranslator = NewTrafficSplitTranslator()
		in = input.NewInputSnapshotManualBuilder("").Build()
	})

})
