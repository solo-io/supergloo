package mesh_workload_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discoveryv1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	controller2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core/mocks"
	mock_core "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/service-mesh-hub/services/common/cluster/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/mocks"
	mock_controllers "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/multicluster/controllers/mocks"
	test_logging "github.com/solo-io/service-mesh-hub/test/logging"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	corev1 "k8s.io/api/core/v1"
	k8s_errs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ = Describe("MeshWorkloadFinder", func() {
	var (
		ctrl                        *gomock.Controller
		ctx                         context.Context
		mockLocalMeshWorkloadClient *mock_core.MockMeshWorkloadClient
		mockLocalMeshClient         *mock_core.MockMeshClient
		mockMeshWorkloadScanner     *mock_mesh_workload.MockMeshWorkloadScanner
		clusterName                 = "clusterName"
		meshWorkloadFinder          mesh_workload.MeshWorkloadFinder
		pod                         = &corev1.Pod{}
		discoveredMeshWorkload      *discoveryv1alpha1.MeshWorkload
		testLogger                  = test_logging.NewTestLogger()
		notFoundErr                 = k8s_errs.NewNotFound(schema.GroupResource{}, "test-not-found-err")
		podClient                   *mock_kubernetes_core.MockPodClient
		podController               *mock_controllers.MockPodController
		meshController              *mock_zephyr_discovery.MockMeshController
		podEventHandlerFuncs        *controller.PodEventHandlerFuncs
		meshEventHandlerFuncs       *controller2.MeshEventHandlerFuncs
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
		mockLocalMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockLocalMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshWorkloadScanner = mock_mesh_workload.NewMockMeshWorkloadScanner(ctrl)
		podClient = mock_kubernetes_core.NewMockPodClient(ctrl)
		podController = mock_controllers.NewMockPodController(ctrl)
		meshController = mock_zephyr_discovery.NewMockMeshController(ctrl)

		podController.EXPECT().
			AddEventHandler(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, h *controller.PodEventHandlerFuncs, predicates ...predicate.Predicate) error {
				podEventHandlerFuncs = h
				return nil
			})

		meshController.EXPECT().
			AddEventHandler(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, h *controller2.MeshEventHandlerFuncs, predicates ...predicate.Predicate) error {
				meshEventHandlerFuncs = h
				return nil
			})

		meshWorkloadFinder = mesh_workload.NewMeshWorkloadFinder(
			ctx,
			clusterName,
			mockLocalMeshWorkloadClient,
			mockLocalMeshClient,
			mesh_workload.MeshWorkloadScannerImplementations{
				core_types.MeshType_ISTIO: mockMeshWorkloadScanner,
			},
			podClient,
		)
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

		err := meshWorkloadFinder.StartDiscovery(podController, meshController)
		Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should create MeshWorkload if Istio injected workload is discovered simultaneously with Istio control plane", func() {
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

		podCopy := *pod
		podCopy.ClusterName = clusterName
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, &podCopy).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		discoveredMeshWorkload.Labels = map[string]string{
			constants.DISCOVERED_BY:             constants.MESH_WORKLOAD_DISCOVERY,
			constants.CLUSTER:                   clusterName,
			constants.KUBE_CONTROLLER_NAME:      discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetName(),
			constants.KUBE_CONTROLLER_NAMESPACE: discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetNamespace(),
		}
		objKey, _ := client.ObjectKeyFromObject(discoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().
			Get(ctx, objKey).
			Return(nil, notFoundErr)
		mockLocalMeshClient.EXPECT().
			List(ctx, &client.ListOptions{}).
			Return(meshList, nil)
		discoveredMeshWorkload.Spec.Mesh = meshSpec
		mockLocalMeshWorkloadClient.EXPECT().
			Create(ctx, discoveredMeshWorkload).
			Return(nil)

		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{*pod}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create MeshWorkload if Istio injected workload found later after Istio is discovered", func() {
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

		podCopy := *pod
		podCopy.ClusterName = clusterName
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, &podCopy).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		discoveredMeshWorkload.Labels = map[string]string{
			constants.DISCOVERED_BY:             constants.MESH_WORKLOAD_DISCOVERY,
			constants.CLUSTER:                   clusterName,
			constants.KUBE_CONTROLLER_NAME:      discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetName(),
			constants.KUBE_CONTROLLER_NAMESPACE: discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetNamespace(),
		}
		objKey, _ := client.ObjectKeyFromObject(discoveredMeshWorkload)
		mockLocalMeshWorkloadClient.EXPECT().
			Get(ctx, objKey).
			Return(nil, notFoundErr)
		mockLocalMeshClient.EXPECT().
			List(ctx, &client.ListOptions{}).
			Return(meshList, nil)
		discoveredMeshWorkload.Spec.Mesh = meshSpec
		mockLocalMeshWorkloadClient.EXPECT().
			Create(ctx, discoveredMeshWorkload).
			Return(nil)

		// no pods when Istio is first discovered
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)

		// Now Istio has been discovered, but no pods will be found yet
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).ToNot(HaveOccurred())

		// now the pod is created separately
		err = podEventHandlerFuncs.OnCreate(pod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("discovers no workload if no mesh has been discovered (prevents a race condition)", func() {
		// no mesh has been discovered
		err := podEventHandlerFuncs.OnCreate(pod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return error if fatal error while scanning workloads", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, metav1.ObjectMeta{}, expectedErr)
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{*pod}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})

		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
	})

	It("should return error if fatal error while populating Mesh resource ref", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(nil, expectedErr)
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{*pod}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).To(HaveOccurred())
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
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
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
		Expect(pod.ClusterName).To(Equal(clusterName))
		Expect(newPod.ClusterName).To(Equal(clusterName))
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
		mockLocalMeshClient.EXPECT().List(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
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
			List(ctx, &client.ListOptions{}).
			Return(meshList, nil).
			Times(2)
		mockLocalMeshWorkloadClient.EXPECT().
			Update(ctx, newDiscoveredMeshWorkload).
			Return(nil)
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return error if error processing old pod", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, metav1.ObjectMeta{}, expectedErr)
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, &corev1.Pod{})
		Expect(err).To(HaveOccurred())
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
	})

	It("should return error if error processing new pod", func() {
		newPod := &corev1.Pod{}
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, metav1.ObjectMeta{}, nil)
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, newPod).Return(nil, metav1.ObjectMeta{}, expectedErr)
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).To(HaveOccurred())
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
	})

	It("should return nil if old and new Pods are not mesh injected", func() {
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, metav1.ObjectMeta{}, nil).Times(2)
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, &corev1.Pod{})
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
			Get(ctx, objKey).
			Return(nil, notFoundErr)
		discoveredMeshWorkload.Spec.Mesh = meshSpec
		discoveredMeshWorkload.Labels = map[string]string{
			constants.DISCOVERED_BY:             constants.MESH_WORKLOAD_DISCOVERY,
			constants.CLUSTER:                   clusterName,
			constants.KUBE_CONTROLLER_NAME:      discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetName(),
			constants.KUBE_CONTROLLER_NAMESPACE: discoveredMeshWorkload.Spec.KubeController.KubeControllerRef.GetNamespace(),
		}
		mockLocalMeshWorkloadClient.EXPECT().
			Create(ctx, discoveredMeshWorkload).
			Return(nil)
		mockLocalMeshClient.EXPECT().
			List(ctx, &client.ListOptions{}).
			Return(meshList, nil)
		podClient.EXPECT().
			List(ctx).
			Return(&corev1.PodList{Items: []corev1.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&discoveryv1alpha1.Mesh{
			Spec: discovery_types.MeshSpec{
				MeshType: &discovery_types.MeshSpec_Istio{},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})
})
