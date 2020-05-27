package appmesh_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/service-mesh-hub/pkg/access-control/enforcer/appmesh"
	mock_appmesh2 "github.com/solo-io/service-mesh-hub/pkg/access-control/enforcer/appmesh/mocks"
	mock_appmesh "github.com/solo-io/service-mesh-hub/pkg/aws/appmesh/mocks"
)

var _ = Describe("AppmeshEnforcer", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   context.Context
		mockAppmeshTranslator *mock_appmesh.MockAppmeshTranslator
		mockDao               *mock_appmesh2.MockAppmeshAccessControlDao
		appmeshEnforcer       appmesh.AppmeshEnforcer
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockAppmeshTranslator = mock_appmesh.NewMockAppmeshTranslator(ctrl)
		mockDao = mock_appmesh2.NewMockAppmeshAccessControlDao(ctrl)
		appmeshEnforcer = appmesh.NewAppmeshEnforcer(mockAppmeshTranslator, mockDao)
	})

	AfterEach(func() {
		ctrl.Finish()
	})
})
