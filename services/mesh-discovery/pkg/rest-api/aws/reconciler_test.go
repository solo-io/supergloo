package aws_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	rest_api "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/rest-api/aws"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
)

var _ = Describe("Reconciler", func() {
	var (
		ctrl                   *gomock.Controller
		ctx                    context.Context
		mockMeshClient         *mock_core.MockMeshClient
		mockMeshWorkloadClient *mock_core.MockMeshWorkloadClient
		mockMeshServiceClient  *mock_core.MockMeshServiceClient
		//mockAppMeshClient      *appmesh.AppMesh
		meshPlatformName           string
		appMeshDiscoveryReconciler rest_api.RestAPIDiscoveryReconciler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		meshPlatformName = "aws-account-name"
		appMeshDiscoveryReconciler = aws.NewAppMeshDiscoveryReconciler(
			mockMeshClient,
			mockMeshWorkloadClient,
			mockMeshServiceClient,
			nil,
			meshPlatformName,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})
})
