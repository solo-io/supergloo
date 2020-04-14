package mesh_workload_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/mocks"
	test_logging "github.com/solo-io/service-mesh-hub/test/logging"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8s_errs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MeshWorkloadFinder", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		mockLocalMeshWorkloadClient *mock_core.MockMeshWorkloadClient
		mockLocalMeshClient         *mock_core.MockMeshClient
		mockMeshWorkloadScanner     *mock_mesh_workload.MockMeshWorkloadScanner
		clusterName                 = "clusterName"
		meshWorkloadFinder          controller.PodEventHandler
		pod                         = &corev1.Pod{}
		discoveredMeshWorkload      *discoveryv1alpha1.MeshWorkload
		testLogger                  = test_logging.NewTestLogger()
		notFoundErr                 = k8s_errs.NewNotFound(schema.GroupResource{}, "test-not-found-err")
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
		mockLocalMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockLocalMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshWorkloadScanner = mock_mesh_workload.NewMockMeshWorkloadScanner(ctrl)

		meshWorkloadFinder = mesh_workload.NewMeshWorkloadFinder(
			ctx,
			clusterName,
			mockLocalMeshWorkloadClient,
			mockLocalMeshClient,
			[]mesh_workload.MeshWorkloadScanner{mockMeshWorkloadScanner})
		discoveredMeshWorkload = &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &core_types.ResourceRef{
						Namespace: "controller-namespace",
					},
				},
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should create MeshWorkload if Istio injected workload found", func() {
		meshName := "meshName"
		meshNamespace := "meshNamespace"
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: meshName, Namespace: meshNamespace},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		meshSpec := &core_types.ResourceRef{
			Name:      mesh.Name,
			Namespace: meshNamespace,
			Cluster:   clusterName,
		}
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		discoveredMeshWorkload.Labels = map[string]string{
			constants.DISCOVERED_BY:             constants.MESH_WORKLOAD_DISCOVERY,
			constants.CLUSTER:                   clusterName,
			constants.KUBE_CONTROLLER_NAME:      discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetName(),
			constants.KUBE_CONTROLLER_NAMESPACE: discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetNamespace(),
		}
		objKey, _ := client.ObjectKeyFromObject(discoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().
			GetMeshWorkload(ctx, objKey).
			Return(nil, notFoundErr)
		mockLocalMeshClient.EXPECT().
			ListMesh(ctx, &client.ListOptions{}).
			Return(meshList, nil)
		discoveredMeshWorkload.Spec.Mesh = meshSpec
		mockLocalMeshWorkloadClient.EXPECT().
			CreateMeshWorkload(ctx, discoveredMeshWorkload).
			Return(nil)

		err := meshWorkloadFinder.CreatePod(pod)

		Expect(pod.ClusterName).To(Equal(clusterName))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return error if fatal error while scanning workloads", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, metav1.ObjectMeta{}, expectedErr)
		err := meshWorkloadFinder.CreatePod(pod)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(HaveLen(1))
		Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
	})

	It("should log warning if non-fatal error while scanning workloads", func() {
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshSpec := &core_types.ResourceRef{
			Name:      mesh.Name,
			Namespace: mesh.Namespace,
			Cluster:   clusterName,
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, expectedErr)
		discoveredMeshWorkload.Labels = map[string]string{
			constants.DISCOVERED_BY:             constants.MESH_WORKLOAD_DISCOVERY,
			constants.CLUSTER:                   clusterName,
			constants.KUBE_CONTROLLER_NAME:      discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetName(),
			constants.KUBE_CONTROLLER_NAMESPACE: discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetNamespace(),
		}
		objKey, _ := client.ObjectKeyFromObject(discoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().
			GetMeshWorkload(ctx, objKey).
			Return(nil, notFoundErr)
		mockLocalMeshClient.EXPECT().
			ListMesh(ctx, &client.ListOptions{}).
			Return(meshList, nil)
		discoveredMeshWorkload.Spec.Mesh = meshSpec
		mockLocalMeshWorkloadClient.EXPECT().
			CreateMeshWorkload(ctx, discoveredMeshWorkload).
			Return(nil)
		_ = meshWorkloadFinder.CreatePod(pod)
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingNonFatal))
	})

	It("should return error if fatal error while populating Mesh resource ref", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		mockLocalMeshClient.EXPECT().ListMesh(ctx, &client.ListOptions{}).Return(nil, expectedErr)
		err := meshWorkloadFinder.CreatePod(pod)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(HaveLen(1))
		Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expectedErr))
		testLogger.EXPECT().LastEntry().HaveMessage(mesh_workload.MeshWorkloadProcessingError)
	})

	It("should create new MeshWorkload if Istio injected workload updated", func() {
		newDiscoveredMeshWorkload := &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &core_types.ResourceRef{
						Namespace: "controller-namespace",
					},
				},
			},
		}
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, newPod).
			Return(newDiscoveredMeshWorkload.Spec.KubeController.KubeControllerRef, newDiscoveredMeshWorkload.ObjectMeta, nil)
		mockLocalMeshClient.EXPECT().ListMesh(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		err := meshWorkloadFinder.UpdatePod(pod, newPod)
		Expect(pod.ClusterName).To(Equal(clusterName))
		Expect(newPod.ClusterName).To(Equal(clusterName))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should do nothing if MeshWorkload unchanged", func() {
		newDiscoveredMeshWorkload := &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &core_types.ResourceRef{
						Namespace: "controller-namespace",
					},
				},
			},
		}
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, newPod).
			Return(newDiscoveredMeshWorkload.Spec.KubeController.KubeControllerRef, newDiscoveredMeshWorkload.ObjectMeta, nil)
		mockLocalMeshClient.EXPECT().ListMesh(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		err := meshWorkloadFinder.UpdatePod(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should update MeshWorkload if changed", func() {
		newNamespace := "new-controller-namespace"
		newDiscoveredMeshWorkload := &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &core_types.ResourceRef{
						Namespace: newNamespace,
					},
				},
			},
		}
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshSpec := &core_types.ResourceRef{
			Name:      mesh.Name,
			Namespace: mesh.Namespace,
			Cluster:   clusterName,
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, newPod).
			Return(newDiscoveredMeshWorkload.Spec.KubeController.KubeControllerRef, newDiscoveredMeshWorkload.ObjectMeta, nil)
		newDiscoveredMeshWorkload.Labels = map[string]string{
			constants.DISCOVERED_BY:             constants.MESH_WORKLOAD_DISCOVERY,
			constants.CLUSTER:                   clusterName,
			constants.KUBE_CONTROLLER_NAME:      newDiscoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetName(),
			constants.KUBE_CONTROLLER_NAMESPACE: newDiscoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetNamespace(),
		}
		newDiscoveredMeshWorkload.Spec.Mesh = meshSpec
		mockLocalMeshClient.EXPECT().
			ListMesh(ctx, &client.ListOptions{}).
			Return(meshList, nil).
			Times(2)
		mockLocalMeshWorkloadClient.EXPECT().
			UpdateMeshWorkload(ctx, newDiscoveredMeshWorkload).
			Return(nil)
		err := meshWorkloadFinder.UpdatePod(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return error if error processing old pod", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, metav1.ObjectMeta{}, expectedErr)
		err := meshWorkloadFinder.UpdatePod(pod, &corev1.Pod{})
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(HaveLen(1))
		Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expectedErr))
		testLogger.EXPECT().LastEntry().HaveMessage(mesh_workload.MeshWorkloadProcessingError)
	})

	It("should return error if error processing new pod", func() {
		newPod := &corev1.Pod{}
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, metav1.ObjectMeta{}, nil)
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, newPod).Return(nil, metav1.ObjectMeta{}, expectedErr)
		err := meshWorkloadFinder.UpdatePod(pod, newPod)
		multierr, ok := err.(*multierror.Error)
		Expect(ok).To(BeTrue())
		Expect(multierr.Errors).To(HaveLen(1))
		Expect(multierr.Errors[0]).To(testutils.HaveInErrorChain(expectedErr))
		testLogger.EXPECT().LastEntry().HaveMessage(mesh_workload.MeshWorkloadProcessingError)
	})

	It("should return nil if old and new Pods are not mesh injected", func() {
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, metav1.ObjectMeta{}, nil).Times(2)
		err := meshWorkloadFinder.UpdatePod(pod, &corev1.Pod{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create new MeshWorkload if pod is now Istio injected", func() {
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshSpec := &core_types.ResourceRef{
			Name:      mesh.Name,
			Namespace: mesh.Namespace,
			Cluster:   clusterName,
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(nil, metav1.ObjectMeta{}, nil)
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		objKey, _ := client.ObjectKeyFromObject(discoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().
			GetMeshWorkload(ctx, objKey).
			Return(nil, notFoundErr)
		discoveredMeshWorkload.Spec.Mesh = meshSpec
		discoveredMeshWorkload.Labels = map[string]string{
			constants.DISCOVERED_BY:             constants.MESH_WORKLOAD_DISCOVERY,
			constants.CLUSTER:                   clusterName,
			constants.KUBE_CONTROLLER_NAME:      discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetName(),
			constants.KUBE_CONTROLLER_NAMESPACE: discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetNamespace(),
		}
		mockLocalMeshWorkloadClient.EXPECT().
			CreateMeshWorkload(ctx, discoveredMeshWorkload).
			Return(nil)
		mockLocalMeshClient.EXPECT().
			ListMesh(ctx, &client.ListOptions{}).
			Return(meshList, nil)
		err := meshWorkloadFinder.UpdatePod(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should log warnings for non-fatal errors on update", func() {
		newDiscoveredMeshWorkload := &discoveryv1alpha1.MeshWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: discovery_types.MeshWorkloadSpec{
				KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &core_types.ResourceRef{
						Namespace: "controller-namespace",
					},
				},
			},
		}
		newPod := &corev1.Pod{}
		mesh := discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				Cluster: &core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: metav1.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &discoveryv1alpha1.MeshList{Items: []discoveryv1alpha1.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, errors.New(""))
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, newPod).
			Return(newDiscoveredMeshWorkload.Spec.KubeController.KubeControllerRef, newDiscoveredMeshWorkload.ObjectMeta, errors.New(""))
		mockLocalMeshClient.EXPECT().ListMesh(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		err := meshWorkloadFinder.UpdatePod(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
		testLogger.EXPECT().LastEntry().HaveMessage(mesh_workload.MeshWorkloadProcessingNonFatal)
	})
})
