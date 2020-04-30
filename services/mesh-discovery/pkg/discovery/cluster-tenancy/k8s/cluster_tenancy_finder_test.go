package k8s_tenancy_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s"
	mock_k8s_tenancy "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/cluster-tenancy/k8s/mocks"
	mock_controllers "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/mesh-platform/event-watcher-factories/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_k8s_core_clients "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
)

var _ = Describe("ClusterTenancyFinder", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   = context.TODO()
		clusterName           = "test-cluster-name"
		mockTenancyScanner    *mock_k8s_tenancy.MockClusterTenancyScanner
		mockPodClient         *mock_k8s_core_clients.MockPodClient
		mockMeshClient        *mock_core.MockMeshClient
		mockPodEventWatcher   *mock_controllers.MockPodEventWatcher
		mockMeshEventWatcher  *mock_zephyr_discovery.MockMeshEventWatcher
		podEventHandlerFuncs  k8s_core_controller.PodEventHandler
		meshEventHandlerFuncs zephyr_discovery_controller.MeshEventHandler
		clusterTenancyFinder  k8s_tenancy.ClusterTenancyFinder
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockTenancyScanner = mock_k8s_tenancy.NewMockClusterTenancyScanner(ctrl)
		mockPodClient = mock_k8s_core_clients.NewMockPodClient(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockPodEventWatcher = mock_controllers.NewMockPodEventWatcher(ctrl)
		mockMeshEventWatcher = mock_zephyr_discovery.NewMockMeshEventWatcher(ctrl)
		clusterTenancyFinder = k8s_tenancy.NewClusterTenancyFinder(
			clusterName,
			[]k8s_tenancy.ClusterTenancyScanner{mockTenancyScanner},
			mockPodClient,
			mockMeshClient,
		)
		mockPodEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandlerFuncs k8s_core_controller.PodEventHandler) error {
				podEventHandlerFuncs = eventHandlerFuncs
				return nil
			})
		mockMeshEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, eventHandlerFuncs zephyr_discovery_controller.MeshEventHandler) error {
				meshEventHandlerFuncs = eventHandlerFuncs
				return nil
			})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should reconcile tenancy upon Pod upsert", func() {
		podEventHandlerFuncs.CreatePod()
	})

})
