package linkerd_test

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/linkerd"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/k8s/mocks"
	k8s_apps_types "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MeshWorkloadScanner", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		mockOwnerFetcher    *mock_mesh_workload.MockOwnerFetcher
		meshWorkloadScanner k8s.MeshWorkloadScanner
		mockMeshClient      *mock_core.MockMeshClient
		namespace           = "namespace"
		clusterName         = "clusterName"
		deploymentName      = "deployment-name"
		deploymentKind      = "deployment-kind"
		pod                 = &k8s_core_types.Pod{
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{Name: "linkerd-proxy"},
				},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: namespace},
		}
		deployment = &k8s_apps_types.Deployment{
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: deploymentName, Namespace: namespace},
			TypeMeta:   k8s_meta_types.TypeMeta{Kind: deploymentKind},
		}
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockOwnerFetcher = mock_mesh_workload.NewMockOwnerFetcher(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		meshWorkloadScanner = linkerd.NewLinkerdMeshWorkloadScanner(mockOwnerFetcher, mockMeshClient, nil)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should scan pod", func() {
		mockOwnerFetcher.EXPECT().GetDeployment(ctx, pod).Return(deployment, nil)
		meshList := &zephyr_discovery.MeshList{
			Items: []zephyr_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh-name",
						Namespace: "mesh-namespace",
					},
					Spec: zephyr_discovery_types.MeshSpec{
						Cluster: &zephyr_core_types.ResourceRef{Name: clusterName},
						MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{
							Linkerd: &zephyr_discovery_types.MeshSpec_LinkerdMesh{},
						},
					},
				},
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(meshList, nil)
		expectedMeshWorkload := &zephyr_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      fmt.Sprintf("linkerd-%s-%s-%s", deploymentName, namespace, clusterName),
				Namespace: container_runtime.GetWriteNamespace(),
				Labels:    linkerd.DiscoveryLabels(),
			},
			Spec: zephyr_discovery_types.MeshWorkloadSpec{
				KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &zephyr_core_types.ResourceRef{
						Name:      deployment.Name,
						Namespace: deployment.Namespace,
						Cluster:   clusterName,
					},
					Labels:             nil,
					ServiceAccountName: "",
				},
				Mesh: &zephyr_core_types.ResourceRef{
					Name:      meshList.Items[0].GetName(),
					Namespace: meshList.Items[0].GetNamespace(),
					Cluster:   clusterName,
				},
			},
		}
		meshWorkload, err := meshWorkloadScanner.ScanPod(ctx, pod, clusterName)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshWorkload).To(Equal(expectedMeshWorkload))
	})

	It("should return nil if not linkerd injected pod", func() {
		nonLinkerdPod := &k8s_core_types.Pod{
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{Image: "random-image"},
				},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Namespace: namespace},
		}
		meshWorkload, err := meshWorkloadScanner.ScanPod(ctx, nonLinkerdPod, clusterName)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshWorkload).To(BeNil())
	})

	It("should return error if error fetching deployment", func() {
		expectedErr := eris.New("error")
		mockOwnerFetcher.EXPECT().GetDeployment(ctx, pod).Return(nil, expectedErr)
		_, err := meshWorkloadScanner.ScanPod(ctx, pod, clusterName)
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
	})
})
