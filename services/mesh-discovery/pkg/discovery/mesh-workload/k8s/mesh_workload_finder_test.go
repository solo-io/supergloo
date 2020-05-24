package k8s_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	mock_controllers "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/compute-target/event-watcher-factories/mocks"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	k8s_core "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ = Describe("MeshWorkloadFinder", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		mockLocalMeshClient         *mock_core.MockMeshClient
		mockLocalMeshWorkloadClient *mock_core.MockMeshWorkloadClient
		mockMeshWorkloadScanner     *mock_mesh_workload.MockMeshWorkloadScanner
		clusterName                 = "clusterName"
		meshWorkloadFinder          k8s.MeshWorkloadFinder
		mockPodClient               *mock_kubernetes_core.MockPodClient
		mockPodEventWatcher         *mock_controllers.MockPodEventWatcher
		mockMeshEventWatcher        *mock_zephyr_discovery.MockMeshEventWatcher
		podEventHandlerFuncs        *k8s_core_controller.PodEventHandlerFuncs
		meshEventHandlerFuncs       *zephyr_discovery_controller.MeshEventHandlerFuncs
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockLocalMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockLocalMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockMeshWorkloadScanner = mock_mesh_workload.NewMockMeshWorkloadScanner(ctrl)
		mockPodClient = mock_kubernetes_core.NewMockPodClient(ctrl)
		mockPodEventWatcher = mock_controllers.NewMockPodEventWatcher(ctrl)
		mockMeshEventWatcher = mock_zephyr_discovery.NewMockMeshEventWatcher(ctrl)

		mockPodEventWatcher.EXPECT().
			AddEventHandler(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, h *k8s_core_controller.PodEventHandlerFuncs, predicates ...predicate.Predicate) error {
				podEventHandlerFuncs = h
				return nil
			})

		mockMeshEventWatcher.EXPECT().
			AddEventHandler(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, h *zephyr_discovery_controller.MeshEventHandlerFuncs, predicates ...predicate.Predicate) error {
				meshEventHandlerFuncs = h
				return nil
			})
		meshWorkloadFinder = k8s.NewMeshWorkloadFinder(
			ctx,
			clusterName,
			mockLocalMeshClient,
			mockLocalMeshWorkloadClient,
			k8s.MeshWorkloadScannerImplementations{
				zephyr_core_types.MeshType_ISTIO1_5: mockMeshWorkloadScanner,
			},
			mockPodClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var attachWorkloadDiscoveryLabels = func(workload *zephyr_discovery.MeshWorkload) {
		workload.Labels = map[string]string{
			constants.DISCOVERED_BY:             constants.MESH_WORKLOAD_DISCOVERY,
			constants.COMPUTE_TARGET:            clusterName,
			constants.KUBE_CONTROLLER_NAME:      workload.Spec.GetKubeController().GetKubeControllerRef().GetName(),
			constants.KUBE_CONTROLLER_NAMESPACE: workload.Spec.GetKubeController().GetKubeControllerRef().GetNamespace(),
		}
	}

	var expectReconcile = func() {
		meshList := &zephyr_discovery.MeshList{Items: []zephyr_discovery.Mesh{
			{Spec: zephyr_discovery_types.MeshSpec{MeshType: &zephyr_discovery_types.MeshSpec_Istio1_5_{}, Cluster: &zephyr_core_types.ResourceRef{Name: clusterName}}},
			{Spec: zephyr_discovery_types.MeshSpec{MeshType: &zephyr_discovery_types.MeshSpec_AwsAppMesh_{}}},
		}}
		mockLocalMeshClient.EXPECT().ListMesh(ctx).Return(meshList, nil)

		extantMeshWorkloadList := &zephyr_discovery.MeshWorkloadList{Items: []zephyr_discovery.MeshWorkload{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload1", Namespace: container_runtime.GetWriteNamespace()}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload2", Namespace: container_runtime.GetWriteNamespace()}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload3", Namespace: container_runtime.GetWriteNamespace()}},
		}}
		mockLocalMeshWorkloadClient.
			EXPECT().
			ListMeshWorkload(ctx, client.MatchingLabels{
				constants.COMPUTE_TARGET: clusterName,
			}).
			Return(extantMeshWorkloadList, nil)

		podList := &k8s_core.PodList{Items: []k8s_core.Pod{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "pod1", Namespace: container_runtime.GetWriteNamespace()}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "pod2", Namespace: container_runtime.GetWriteNamespace()}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "pod4", Namespace: container_runtime.GetWriteNamespace()}},
		}}
		discoveredMeshWorkloads := []*zephyr_discovery.MeshWorkload{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload1", Namespace: container_runtime.GetWriteNamespace()},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{Mesh: &zephyr_core_types.ResourceRef{Name: "mesh"}}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload2", Namespace: container_runtime.GetWriteNamespace()}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload4", Namespace: container_runtime.GetWriteNamespace()}},
		}
		for _, workload := range discoveredMeshWorkloads {
			attachWorkloadDiscoveryLabels(workload)
		}
		mockPodClient.EXPECT().ListPod(ctx).Return(podList, nil)
		for i, pod := range podList.Items {
			pod := pod
			mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, &pod, clusterName).Return(discoveredMeshWorkloads[i], nil)
		}

		// workload1 should be updated
		mockLocalMeshWorkloadClient.EXPECT().UpsertMeshWorkloadSpec(ctx, discoveredMeshWorkloads[0]).Return(nil)
		// workload3 should be deleted
		mockLocalMeshWorkloadClient.
			EXPECT().
			DeleteMeshWorkload(ctx, clients.ObjectMetaToObjectKey(extantMeshWorkloadList.Items[2].ObjectMeta)).
			Return(nil)
		// workload4 should be created
		mockLocalMeshWorkloadClient.EXPECT().UpsertMeshWorkloadSpec(ctx, discoveredMeshWorkloads[2]).Return(nil)
	}

	It("should reconcile MeshWorkloads upon pod create", func() {
		expectReconcile()
		err := meshWorkloadFinder.StartDiscovery(mockPodEventWatcher, mockMeshEventWatcher)
		Expect(err).ToNot(HaveOccurred())
		expectReconcile()
		err = podEventHandlerFuncs.CreatePod(&k8s_core.Pod{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile MeshWorkloads upon pod update", func() {
		expectReconcile()
		err := meshWorkloadFinder.StartDiscovery(mockPodEventWatcher, mockMeshEventWatcher)
		Expect(err).ToNot(HaveOccurred())
		expectReconcile()
		err = podEventHandlerFuncs.UpdatePod(nil, &k8s_core.Pod{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile MeshWorkloads upon pod delete", func() {
		expectReconcile()
		err := meshWorkloadFinder.StartDiscovery(mockPodEventWatcher, mockMeshEventWatcher)
		Expect(err).ToNot(HaveOccurred())
		expectReconcile()
		err = podEventHandlerFuncs.DeletePod(&k8s_core.Pod{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile MeshWorkloads upon mesh create", func() {
		expectReconcile()
		err := meshWorkloadFinder.StartDiscovery(mockPodEventWatcher, mockMeshEventWatcher)
		Expect(err).ToNot(HaveOccurred())
		expectReconcile()
		err = meshEventHandlerFuncs.CreateMesh(&zephyr_discovery.Mesh{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile MeshWorkloads upon mesh update", func() {
		expectReconcile()
		err := meshWorkloadFinder.StartDiscovery(mockPodEventWatcher, mockMeshEventWatcher)
		Expect(err).ToNot(HaveOccurred())
		expectReconcile()
		err = meshEventHandlerFuncs.UpdateMesh(nil, &zephyr_discovery.Mesh{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should reconcile MeshWorkloads upon mesh delete", func() {
		expectReconcile()
		err := meshWorkloadFinder.StartDiscovery(mockPodEventWatcher, mockMeshEventWatcher)
		Expect(err).ToNot(HaveOccurred())
		expectReconcile()
		err = meshEventHandlerFuncs.DeleteMesh(&zephyr_discovery.Mesh{})
		Expect(err).ToNot(HaveOccurred())
	})
})
