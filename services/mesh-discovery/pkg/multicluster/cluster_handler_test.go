package multicluster_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/mesh-projects/pkg/common/concurrency"
	mock_concurrency "github.com/solo-io/mesh-projects/pkg/common/concurrency/mocks"
	mock_docker "github.com/solo-io/mesh-projects/pkg/common/docker/mocks"
	"github.com/solo-io/mesh-projects/services/common"
	mock_mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh"
	mesh_workload "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/mesh-workload"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster"
	mock_controllers "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster/controllers/mocks"
	mock_controller_runtime "github.com/solo-io/mesh-projects/test/mocks/controller-runtime"
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
		deploymentControllerFactory := mock_controllers.NewMockDeploymentControllerFactory(ctrl)
		deploymentController := mock_controllers.NewMockDeploymentController(ctrl)
		podControllerFactory := mock_controllers.NewMockPodControllerFactory(ctrl)
		podController := mock_controllers.NewMockPodController(ctrl)
		localAsyncManager := mock_mc_manager.NewMockAsyncManager(ctrl)
		localManager := mock_controller_runtime.NewMockManager(ctrl)
		localMeshClient := mock_core.NewMockMeshClient(ctrl)
		localMeshWorkloadClient := mock_core.NewMockMeshWorkloadClient(ctrl)
		threadSafeMap := mock_concurrency.NewMockThreadSafeMap(ctrl)
		imageParser := mock_docker.NewMockImageNameParser(ctrl)
		mockDynamicClient := mock_controller_runtime.NewMockClient(ctrl)

		expectedMeshFinder := mesh.DefaultMeshFinder(
			ctx,
			common.LocalClusterName,
			[]mesh.MeshScanner{},
			localMeshClient,
		)

		localAsyncManager.EXPECT().Manager().Return(localManager)
		localManager.EXPECT().GetClient().Return(mockDynamicClient)
		expectedMeshWorkloadFinder := mesh_workload.DefaultMeshWorkloadFinder(
			common.LocalClusterName,
			ctx,
			localMeshWorkloadClient,
			localMeshClient,
			mockDynamicClient,
			imageParser,
		)

		deploymentControllerFactory.EXPECT().Build(localAsyncManager, common.LocalClusterName).
			Return(deploymentController, nil)
		podControllerFactory.EXPECT().Build(localAsyncManager, common.LocalClusterName).
			Return(podController, nil)

		deploymentController.EXPECT().AddEventHandler(ctx, expectedMeshFinder, multicluster.ObjectPredicates)
		podController.EXPECT().AddEventHandler(ctx, expectedMeshWorkloadFinder, multicluster.ObjectPredicates)

		threadSafeMap.EXPECT().Store(common.LocalClusterName, deploymentController)

		clusterHandler, err := multicluster.NewDiscoveryClusterHandler(
			ctx,
			imageParser,
			localAsyncManager,
			deploymentControllerFactory,
			localMeshClient,
			[]mesh.MeshScanner{},
			podControllerFactory,
			localMeshWorkloadClient,
			func() concurrency.ThreadSafeMap { return threadSafeMap },
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(clusterHandler).NotTo(BeNil())
	})
})
