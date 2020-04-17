package istio_test

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/istio"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/mocks"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MeshWorkloadScanner", func() {
	var (
		ctrl                *gomock.Controller
		ctx                 context.Context
		mockOwnerFetcher    *mock_mesh_workload.MockOwnerFetcher
		meshWorkloadScanner mesh_workload.MeshWorkloadScanner
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
			ObjectMeta: metav1.ObjectMeta{ClusterName: clusterName, Namespace: namespace},
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
		meshWorkloadScanner = istio.NewIstioMeshWorkloadScanner(mockOwnerFetcher)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should scan pod", func() {
		expectedMeshWorkload := &zephyr_discovery.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("istio-%s-%s-%s", deploymentName, namespace, clusterName),
				Namespace: env.GetWriteNamespace(),
				Labels:    istio.DiscoveryLabels(),
			},
			Spec: zephyr_discovery_types.MeshWorkloadSpec{
				KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &zephyr_core_types.ResourceRef{
						Name:      deployment.Name,
						Namespace: deployment.Namespace,
						Cluster:   pod.ObjectMeta.ClusterName,
					},
					Labels:             nil,
					ServiceAccountName: "",
				},
			},
		}
		mockOwnerFetcher.EXPECT().GetDeployment(ctx, pod).Return(deployment, nil)
		controllerRef, meta, err := meshWorkloadScanner.ScanPod(ctx, pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(controllerRef).To(Equal(expectedMeshWorkload.Spec.KubeController.KubeControllerRef))
		Expect(meta).To(Equal(expectedMeshWorkload.ObjectMeta))
	})

	It("should return nil if not istio injected pod", func() {
		nonIstioPod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Image: "random-image"},
				},
			},
			ObjectMeta: metav1.ObjectMeta{ClusterName: clusterName, Namespace: namespace},
		}
		meshWorkload, _, err := meshWorkloadScanner.ScanPod(ctx, nonIstioPod)
		Expect(err).NotTo(HaveOccurred())
		Expect(meshWorkload).To(BeNil())
	})

	It("should return error if error fetching deployment", func() {
		expectedErr := eris.New("error")
		mockOwnerFetcher.EXPECT().GetDeployment(ctx, pod).Return(nil, expectedErr)
		_, _, err := meshWorkloadScanner.ScanPod(ctx, pod)
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
	})
})
