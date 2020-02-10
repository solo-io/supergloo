package multicluster_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/mesh-projects/pkg/common/concurrency"
	mock_concurrency "github.com/solo-io/mesh-projects/pkg/common/concurrency/mocks"
	"github.com/solo-io/mesh-projects/services/common"
	mock_mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery"
	mock_discovery "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster"
	mock_multicluster "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster/mocks"
)

var _ = Describe("Local mesh client", func() {
	var (
		ctrl *gomock.Controller
		ctx  = context.TODO()
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("watches the local cluster on startup", func() {
		controllerFactory := mock_multicluster.NewMockDeploymentControllerFactory(ctrl)
		deploymentController := mock_multicluster.NewMockDeploymentController(ctrl)
		localManager := mock_mc_manager.NewMockAsyncManager(ctrl)
		localMeshClient := mock_discovery.NewMockLocalMeshClient(ctrl)
		threadSafeMap := mock_concurrency.NewMockThreadSafeMap(ctrl)
		meshDiscoverer := discovery.NewMeshDiscoverer(
			ctx,
			common.LocalClusterName,
			[]discovery.MeshFinder{},
			localMeshClient,
		)

		controllerFactory.EXPECT().Build(localManager, common.LocalClusterName).
			Return(deploymentController, nil)

		deploymentController.EXPECT().AddEventHandler(ctx, meshDiscoverer, multicluster.DeploymentPredicates)

		threadSafeMap.EXPECT().Store(common.LocalClusterName, deploymentController)

		clusterHandler, err := multicluster.NewMeshDiscoveryClusterHandler(
			ctx,
			controllerFactory,
			localManager,
			localMeshClient,
			[]discovery.MeshFinder{},
			func() concurrency.ThreadSafeMap { return threadSafeMap },
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(clusterHandler).NotTo(BeNil())
	})
})
