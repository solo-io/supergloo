package multicluster_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kubernetes_apps "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/apps"
	mock_kubernetes_apps "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/apps/mocks"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	mock_kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core/mocks"
	discovery_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/mesh-projects/services/common"
	mock_mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh"
	mesh_workload "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh-workload"
	mock_mesh_workload "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/discovery/mesh-workload/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster"
	mock_controllers "github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/multicluster/controllers/mocks"
	"github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/wire"
	mock_controller_runtime "github.com/solo-io/mesh-projects/test/mocks/controller-runtime"
	mock_corev1 "github.com/solo-io/mesh-projects/test/mocks/corev1"
	mock_zephyr_core "github.com/solo-io/mesh-projects/test/mocks/zephyr/core"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Cluster Handler", func() {
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
		localMeshClient := mock_core.NewMockMeshClient(ctrl)
		localMeshWorkloadClient := mock_core.NewMockMeshWorkloadClient(ctrl)
		serviceControllerFactory := mock_controllers.NewMockServiceControllerFactory(ctrl)
		meshWorkloadControllerFactory := mock_controllers.NewMockMeshWorkloadControllerFactory(ctrl)
		serviceController := mock_corev1.NewMockServiceController(ctrl)
		meshWorkloadController := mock_zephyr_core.NewMockMeshWorkloadController(ctrl)
		serviceClient := mock_kubernetes_core.NewMockServiceClient(ctrl)
		meshServiceClient := mock_core.NewMockMeshServiceClient(ctrl)
		meshWorkloadClient := mock_core.NewMockMeshWorkloadClient(ctrl)

		expectedMeshFinder := mesh.NewMeshFinder(
			ctx,
			common.LocalClusterName,
			[]mesh.MeshScanner{},
			localMeshClient,
		)

		meshWorkloadScanner := mock_mesh_workload.NewMockMeshWorkloadScanner(ctrl)

		expectedMeshWorkloadFinder := mesh_workload.NewMeshWorkloadFinder(
			ctx,
			common.LocalClusterName,
			localMeshWorkloadClient,
			localMeshClient,
			[]mesh_workload.MeshWorkloadScanner{meshWorkloadScanner},
		)

		serviceController.EXPECT().AddEventHandler(ctx, gomock.Any()).Return(nil)
		meshWorkloadController.EXPECT().AddEventHandler(ctx, gomock.Any()).Return(nil)

		deploymentControllerFactory.EXPECT().Build(localAsyncManager, common.LocalClusterName).
			Return(deploymentController, nil)
		podControllerFactory.EXPECT().Build(localAsyncManager, common.LocalClusterName).
			Return(podController, nil)
		serviceControllerFactory.EXPECT().Build(localAsyncManager, common.LocalClusterName).
			Return(serviceController, nil)
		meshWorkloadControllerFactory.EXPECT().Build(localAsyncManager, common.LocalClusterName).
			Return(meshWorkloadController, nil)

		deploymentController.EXPECT().AddEventHandler(ctx, expectedMeshFinder, multicluster.MeshPredicates)
		podController.EXPECT().AddEventHandler(ctx, expectedMeshWorkloadFinder, multicluster.MeshWorkloadPredicates)

		deploymentClient := mock_kubernetes_apps.NewMockDeploymentClient(ctrl)
		replicaSetClient := mock_kubernetes_apps.NewMockReplicaSetClient(ctrl)
		ownerFetcherClient := mock_mesh_workload.NewMockOwnerFetcher(ctrl)
		mockK8sManager := mock_controller_runtime.NewMockManager(ctrl)
		dynamicClient := mock_controller_runtime.NewMockClient(ctrl)

		localAsyncManager.EXPECT().Manager().Return(mockK8sManager).Times(2)
		mockK8sManager.EXPECT().GetClient().Return(dynamicClient).Times(2)

		clusterHandler, err := multicluster.NewDiscoveryClusterHandler(
			ctx,
			localAsyncManager,
			localMeshClient,
			[]mesh.MeshScanner{},
			[]mesh_workload.MeshWorkloadScannerFactory{
				func(_ mesh_workload.OwnerFetcher) mesh_workload.IstioMeshWorkloadScanner {
					return meshWorkloadScanner
				},
			},
			localMeshWorkloadClient,
			wire.DiscoveryContext{
				ClientFactories: wire.ClientFactories{
					DeploymentClientFactory: func(_ client.Client) kubernetes_apps.DeploymentClient {
						return deploymentClient
					},
					ReplicaSetClientFactory: func(_ client.Client) kubernetes_apps.ReplicaSetClient {
						return replicaSetClient
					},
					OwnerFetcherClientFactory: func(_ kubernetes_apps.DeploymentClient, _ kubernetes_apps.ReplicaSetClient) mesh_workload.OwnerFetcher {
						return ownerFetcherClient
					},
					ServiceClientFactory: func(_ client.Client) kubernetes_core.ServiceClient {
						return serviceClient
					},
					MeshServiceClientFactory: func(_ client.Client) discovery_core.MeshServiceClient {
						return meshServiceClient
					},
					MeshWorkloadClientFactory: func(_ client.Client) discovery_core.MeshWorkloadClient {
						return meshWorkloadClient
					},
				},
				ControllerFactories: wire.ControllerFactories{
					DeploymentControllerFactory:   deploymentControllerFactory,
					PodControllerFactory:          podControllerFactory,
					ServiceControllerFactory:      serviceControllerFactory,
					MeshWorkloadControllerFactory: meshWorkloadControllerFactory,
				},
			},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(clusterHandler).NotTo(BeNil())
	})
})
