package translation_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/solo-io/service-mesh-hub/pkg/aws/appmesh/translation"
)

var _ = Describe("Translator", func() {
	var (
		ctrl              *gomock.Controller
		appmeshTranslator AppmeshTranslator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		appmeshTranslator = NewAppmeshTranslator()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should build VirtualNode", func() {

	})
})
