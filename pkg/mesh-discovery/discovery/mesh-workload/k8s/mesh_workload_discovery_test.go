package k8s_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_kubernetes_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s/mocks"
	k8s_core "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MeshWorkloadDiscovery", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		mockLocalMeshClient         *mock_core.MockMeshClient
		mockLocalMeshWorkloadClient *mock_core.MockMeshWorkloadClient
		mockMeshWorkloadScanner     *mock_mesh_workload.MockMeshWorkloadScanner
		clusterName                 = "clusterName"
		meshWorkloadDiscovery       k8s.MeshWorkloadDiscovery
		mockMulticlusterClientset   *mock_kubernetes_core.MockMulticlusterClientset
		mockClientset               *mock_kubernetes_core.MockClientset
		mockPodClient               *mock_kubernetes_core.MockPodClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockLocalMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockLocalMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockMeshWorkloadScanner = mock_mesh_workload.NewMockMeshWorkloadScanner(ctrl)
		mockMulticlusterClientset = mock_kubernetes_core.NewMockMulticlusterClientset(ctrl)
		mockClientset = mock_kubernetes_core.NewMockClientset(ctrl)
		mockPodClient = mock_kubernetes_core.NewMockPodClient(ctrl)
		meshWorkloadDiscovery = k8s.NewMeshWorkloadDiscovery(
			mockLocalMeshClient,
			mockLocalMeshWorkloadClient,
			k8s.MeshWorkloadScannerImplementations{
				smh_core_types.MeshType_ISTIO1_5: mockMeshWorkloadScanner,
			},
			mockMulticlusterClientset,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var attachWorkloadDiscoveryLabels = func(workload *smh_discovery.MeshWorkload) {
		workload.Labels = map[string]string{
			kube.DISCOVERED_BY:             kube.MESH_WORKLOAD_DISCOVERY,
			kube.COMPUTE_TARGET:            clusterName,
			kube.KUBE_CONTROLLER_NAME:      workload.Spec.GetKubeController().GetKubeControllerRef().GetName(),
			kube.KUBE_CONTROLLER_NAMESPACE: workload.Spec.GetKubeController().GetKubeControllerRef().GetNamespace(),
		}
	}

	var expectReconcile = func() {
		mockMulticlusterClientset.EXPECT().Cluster(clusterName).Return(mockClientset, nil)
		mockClientset.EXPECT().Pods().Return(mockPodClient)

		meshList := &smh_discovery.MeshList{Items: []smh_discovery.Mesh{
			{Spec: smh_discovery_types.MeshSpec{MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{}, Cluster: &smh_core_types.ResourceRef{Name: clusterName}}},
			{Spec: smh_discovery_types.MeshSpec{MeshType: &smh_discovery_types.MeshSpec_AwsAppMesh_{}}},
		}}
		mockLocalMeshClient.EXPECT().ListMesh(ctx).Return(meshList, nil)

		extantMeshWorkloadList := &smh_discovery.MeshWorkloadList{Items: []smh_discovery.MeshWorkload{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload1", Namespace: container_runtime.GetWriteNamespace()}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload2", Namespace: container_runtime.GetWriteNamespace()}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload3", Namespace: container_runtime.GetWriteNamespace()}},
		}}
		mockLocalMeshWorkloadClient.
			EXPECT().
			ListMeshWorkload(ctx, client.MatchingLabels{
				kube.COMPUTE_TARGET: clusterName,
			}).
			Return(extantMeshWorkloadList, nil)

		podList := &k8s_core.PodList{Items: []k8s_core.Pod{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "pod1", Namespace: container_runtime.GetWriteNamespace()}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "pod2", Namespace: container_runtime.GetWriteNamespace()}},
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "pod4", Namespace: container_runtime.GetWriteNamespace()}},
		}}
		discoveredMeshWorkloads := []*smh_discovery.MeshWorkload{
			{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "workload1", Namespace: container_runtime.GetWriteNamespace()},
				Spec: smh_discovery_types.MeshWorkloadSpec{Mesh: &smh_core_types.ResourceRef{Name: "mesh"}}},
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
		mockLocalMeshWorkloadClient.EXPECT().UpsertMeshWorkload(ctx, discoveredMeshWorkloads[0]).Return(nil)
		// workload3 should be deleted
		mockLocalMeshWorkloadClient.
			EXPECT().
			DeleteMeshWorkload(ctx, selection.ObjectMetaToObjectKey(extantMeshWorkloadList.Items[2].ObjectMeta)).
			Return(nil)
		// workload4 should be created
		mockLocalMeshWorkloadClient.EXPECT().UpsertMeshWorkload(ctx, discoveredMeshWorkloads[2]).Return(nil)
	}

	It("should discover MeshWorkloads", func() {
		expectReconcile()
		err := meshWorkloadDiscovery.DiscoverMeshWorkloads(ctx, clusterName)
		Expect(err).ToNot(HaveOccurred())
	})
})
