package linkerd_test

import (
	"context"
	"fmt"

	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload/linkerd"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/k8s/mesh-workload/mocks"
	k8s_apps_types "k8s.io/api/apps/v1"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		pod                 = &k8s_core_types.Pod{
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{Name: "linkerd-proxy"},
				},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{ClusterName: clusterName, Namespace: namespace},
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
		meshWorkloadScanner = linkerd.NewLinkerdMeshWorkloadScanner(mockOwnerFetcher)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should scan pod", func() {
		expectedMeshWorkload := &zephyr_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      fmt.Sprintf("linkerd-%s-%s-%s", deploymentName, namespace, clusterName),
				Namespace: env.GetWriteNamespace(),
				Labels:    linkerd.DiscoveryLabels(),
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

	It("should return nil if not linkerd injected pod", func() {
		nonLinkerdPod := &k8s_core_types.Pod{
			Spec: k8s_core_types.PodSpec{
				Containers: []k8s_core_types.Container{
					{Image: "random-image"},
				},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{ClusterName: clusterName, Namespace: namespace},
		}
		meshWorkload, _, err := meshWorkloadScanner.ScanPod(ctx, nonLinkerdPod)
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
