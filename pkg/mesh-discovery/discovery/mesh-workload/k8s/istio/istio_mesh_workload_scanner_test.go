package istio_test

import (
	"context"
	"fmt"

	mock_core "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s/istio"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-workload/k8s/mocks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		pod                 = &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Image: "istio-proxy"},
				},
			},
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace},
		}
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
			TypeMeta:   metav1.TypeMeta{Kind: deploymentKind},
		}
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockOwnerFetcher = mock_mesh_workload.NewMockOwnerFetcher(ctrl)
		mockMeshClient = mock_core.NewMockMeshClient(ctrl)
		meshWorkloadScanner = istio.NewIstioMeshWorkloadScanner(
			func(clusterName string) (k8s.OwnerFetcher, error) {
				return mockOwnerFetcher, nil
			},
			mockMeshClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should scan pod", func() {
		mockOwnerFetcher.EXPECT().GetDeployment(ctx, pod).Return(deployment, nil)
		meshList := &smh_discovery.MeshList{
			Items: []smh_discovery.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mesh-name",
						Namespace: "mesh-namespace",
					},
					Spec: smh_discovery_types.MeshSpec{
						Cluster: &smh_core_types.ResourceRef{Name: clusterName},
						MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{
							Istio1_5: &smh_discovery_types.MeshSpec_Istio1_5{
								Metadata: &smh_discovery_types.MeshSpec_IstioMesh{},
							},
						},
					},
				},
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(meshList, nil)
		expectedMeshWorkload := &smh_discovery.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("istio-%s-%s-%s", deploymentName, namespace, clusterName),
				Namespace: container_runtime.GetWriteNamespace(),
				Labels:    istio.DiscoveryLabels(smh_core_types.MeshType_ISTIO1_5),
			},
			Spec: smh_discovery_types.MeshWorkloadSpec{
				KubeController: &smh_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &smh_core_types.ResourceRef{
						Name:      deployment.Name,
						Namespace: deployment.Namespace,
						Cluster:   clusterName,
					},
					Labels:             nil,
					ServiceAccountName: "",
				},
				Mesh: &smh_core_types.ResourceRef{
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

	It("should work for istio 1.6", func() {
		mockOwnerFetcher.EXPECT().GetDeployment(ctx, pod).Return(deployment, nil)
		meshList := &smh_discovery.MeshList{
			Items: []smh_discovery.Mesh{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mesh-name",
						Namespace: "mesh-namespace",
					},
					Spec: smh_discovery_types.MeshSpec{
						Cluster: &smh_core_types.ResourceRef{Name: clusterName},
						MeshType: &smh_discovery_types.MeshSpec_Istio1_6_{
							Istio1_6: &smh_discovery_types.MeshSpec_Istio1_6{
								Metadata: &smh_discovery_types.MeshSpec_IstioMesh{},
							},
						},
					},
				},
			},
		}
		mockMeshClient.EXPECT().ListMesh(ctx).Return(meshList, nil)
		expectedMeshWorkload := &smh_discovery.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("istio-%s-%s-%s", deploymentName, namespace, clusterName),
				Namespace: container_runtime.GetWriteNamespace(),
				Labels:    istio.DiscoveryLabels(smh_core_types.MeshType_ISTIO1_6),
			},
			Spec: smh_discovery_types.MeshWorkloadSpec{
				KubeController: &smh_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &smh_core_types.ResourceRef{
						Name:      deployment.Name,
						Namespace: deployment.Namespace,
						Cluster:   clusterName,
					},
					Labels:             nil,
					ServiceAccountName: "",
				},
				Mesh: &smh_core_types.ResourceRef{
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

	It("should return nil if not istio injected pod", func() {
		nonIstioPod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Image: "random-image"},
				},
			},
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace},
		}
		meshWorkload, err := meshWorkloadScanner.ScanPod(ctx, nonIstioPod, clusterName)
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
