package k8s_tenancy_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8s_core_controller "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/controller"
	mock_k8s_core_clients "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	mock_controllers "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/event-watcher-factories/mocks"
	k8s_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s"
	mock_k8s_tenancy "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/cluster-tenancy/k8s/mocks"
	mock_smh_discovery "github.com/solo-io/service-mesh-hub/test/mocks/smh/discovery"
	k8s_core "k8s.io/api/core/v1"
	k8s_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ClusterTenancyRegistrarLoop", func() {
	var (
		ctrl                  *gomock.Controller
		ctx                   = context.TODO()
		clusterName           = "test-cluster-name"
		mockTenancyRegistrar  *mock_k8s_tenancy.MockClusterTenancyRegistrar
		mockPodClient         *mock_k8s_core_clients.MockPodClient
		mockMeshClient        *mock_core.MockMeshClient
		mockPodEventWatcher   *mock_controllers.MockPodEventWatcher
		mockMeshEventWatcher  *mock_smh_discovery.MockMeshEventWatcher
		podEventHandlerFuncs  k8s_core_controller.PodEventHandler
		meshEventHandlerFuncs smh_discovery_controller.MeshEventHandler
		clusterTenancyFinder  k8s_tenancy.ClusterTenancyRegistrarLoop
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockTenancyRegistrar = mock_k8s_tenancy.NewMockClusterTenancyRegistrar(ctrl)
		mockPodClient = mock_k8s_core_clients.NewMockPodClient(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockPodEventWatcher = mock_controllers.NewMockPodEventWatcher(ctrl)
		mockMeshEventWatcher = mock_smh_discovery.NewMockMeshEventWatcher(ctrl)
		clusterTenancyFinder = k8s_tenancy.NewClusterTenancyFinder(
			clusterName,
			[]k8s_tenancy.ClusterTenancyRegistrar{mockTenancyRegistrar},
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
			DoAndReturn(func(ctx context.Context, eventHandlerFuncs smh_discovery_controller.MeshEventHandler) error {
				meshEventHandlerFuncs = eventHandlerFuncs
				return nil
			})
		err := clusterTenancyFinder.StartRegistration(ctx, mockPodEventWatcher, mockMeshEventWatcher)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectReconcile = func() {
		namespace := "namespace"
		podList := &k8s_core.PodList{
			Items: []k8s_core.Pod{
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "pod1", Namespace: namespace}},
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "pod2", Namespace: namespace}},
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "pod3", Namespace: namespace}},
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "pod4", Namespace: namespace}},
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "pod5", Namespace: namespace}},
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "pod6", Namespace: namespace}},
			},
		}
		mockPodClient.EXPECT().ListPod(ctx).Return(podList, nil)
		meshList := &smh_discovery.MeshList{
			Items: []smh_discovery.Mesh{
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "mesh1", Namespace: namespace}},
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "mesh2", Namespace: namespace}},
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "mesh3", Namespace: namespace}},
				{ObjectMeta: k8s_meta.ObjectMeta{Name: "mesh4", Namespace: namespace}}, // Not on cluster
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(meshList, nil)
		for i, pod := range podList.Items[:3] {
			pod := pod
			mockTenancyRegistrar.EXPECT().MeshFromSidecar(ctx, &pod).Return(&meshList.Items[i], nil)
		}
		for _, pod := range podList.Items[3:] {
			pod := pod
			mockTenancyRegistrar.EXPECT().MeshFromSidecar(ctx, &pod).Return(nil, nil)
		}
		mockTenancyRegistrar.EXPECT().ClusterHostsMesh(clusterName, &meshList.Items[0]).Return(false)
		mockTenancyRegistrar.EXPECT().RegisterMesh(ctx, clusterName, &meshList.Items[0]).Return(nil)
		mockTenancyRegistrar.EXPECT().ClusterHostsMesh(clusterName, &meshList.Items[1]).Return(true)
		mockTenancyRegistrar.EXPECT().ClusterHostsMesh(clusterName, &meshList.Items[2]).Return(true)
		mockTenancyRegistrar.EXPECT().ClusterHostsMesh(clusterName, &meshList.Items[3]).Return(true)
		mockTenancyRegistrar.EXPECT().DeregisterMesh(ctx, clusterName, &meshList.Items[3]).Return(nil)
	}

	It("should reconcile tenancy upon Pod create", func() {
		expectReconcile()
		err := podEventHandlerFuncs.CreatePod(&k8s_core.Pod{})
		Expect(err).To(BeNil())
	})

	It("should reconcile tenancy upon Pod update", func() {
		expectReconcile()
		err := podEventHandlerFuncs.UpdatePod(&k8s_core.Pod{}, &k8s_core.Pod{})
		Expect(err).To(BeNil())
	})

	It("should reconcile tenancy upon Pod delete", func() {
		expectReconcile()
		err := podEventHandlerFuncs.DeletePod(&k8s_core.Pod{})
		Expect(err).To(BeNil())
	})

	It("should reconcile tenancy upon Mesh create", func() {
		expectReconcile()
		err := meshEventHandlerFuncs.CreateMesh(&smh_discovery.Mesh{})
		Expect(err).To(BeNil())
	})

	It("should reconcile tenancy upon Mesh update", func() {
		expectReconcile()
		err := meshEventHandlerFuncs.UpdateMesh(&smh_discovery.Mesh{}, &smh_discovery.Mesh{})
		Expect(err).To(BeNil())
	})
})
