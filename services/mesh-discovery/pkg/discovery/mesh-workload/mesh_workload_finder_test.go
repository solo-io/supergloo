package mesh_workload_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/testutils"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_controller "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload"
	mock_mesh_workload "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-workload/mocks"
	mock_controllers "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/multicluster/controllers/mocks"
	test_logging "github.com/solo-io/service-mesh-hub/test/logging"
	mock_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_errs "k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		pod                         = &k8s_core_types.Pod{}
		discoveredMeshWorkload      *zephyr_discovery.MeshWorkload
		testLogger                  = test_logging.NewTestLogger()
		notFoundErr                 = k8s_errs.NewNotFound(schema.GroupResource{}, "test-not-found-err")
		podClient                   *mock_kubernetes_core.MockPodClient
		podEventWatcher             *mock_controllers.MockPodEventWatcher
		meshEventWatcher            *mock_zephyr_discovery.MockMeshEventWatcher
		podEventHandlerFuncs        *k8s_core_controller.PodEventHandlerFuncs
		meshEventHandlerFuncs       *zephyr_discovery_controller.MeshEventHandlerFuncs
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = contextutils.WithExistingLogger(context.TODO(), testLogger.Logger())
		mockLocalMeshWorkloadClient = mock_core.NewMockMeshWorkloadClient(ctrl)
		mockLocalMeshClient = mock_core.NewMockMeshClient(ctrl)
		mockMeshWorkloadScanner = mock_mesh_workload.NewMockMeshWorkloadScanner(ctrl)
		podClient = mock_kubernetes_core.NewMockPodClient(ctrl)
		podEventWatcher = mock_controllers.NewMockPodEventWatcher(ctrl)
		meshEventWatcher = mock_zephyr_discovery.NewMockMeshEventWatcher(ctrl)

		podEventWatcher.EXPECT().
			AddEventHandler(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, h *k8s_core_controller.PodEventHandlerFuncs, predicates ...predicate.Predicate) error {
				podEventHandlerFuncs = h
				return nil
			})

		meshEventWatcher.EXPECT().
			AddEventHandler(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, h *zephyr_discovery_controller.MeshEventHandlerFuncs, predicates ...predicate.Predicate) error {
				meshEventHandlerFuncs = h
				return nil
			})

		meshWorkloadFinder = mesh_workload.NewMeshWorkloadFinder(
			ctx,
			clusterName,
			mockLocalMeshWorkloadClient,
			mockLocalMeshClient,
			mesh_workload.MeshWorkloadScannerImplementations{
				zephyr_core_types.MeshType_ISTIO: mockMeshWorkloadScanner,
			},
			podClient,
		)
		discoveredMeshWorkload = &zephyr_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: zephyr_discovery_types.MeshWorkloadSpec{
				KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &zephyr_core_types.ResourceRef{
						Namespace: "controller-namespace",
					},
				},
			},
		}

		err := meshWorkloadFinder.StartDiscovery(podEventWatcher, meshEventWatcher)
		Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should create MeshWorkload if Istio injected workload is discovered simultaneously with Istio control plane", func() {
		meshName := "meshName"
		meshNamespace := "meshNamespace"
		mesh := zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: meshName, Namespace: meshNamespace},
		}
		meshList := &zephyr_discovery.MeshList{Items: []zephyr_discovery.Mesh{mesh}}
		meshSpec := &zephyr_core_types.ResourceRef{
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
			GetMeshWorkload(ctx, objKey).
			Return(nil, notFoundErr)
		mockLocalMeshClient.EXPECT().
			ListMesh(ctx, &client.ListOptions{}).
			Return(meshList, nil)
		discoveredMeshWorkload.Spec.Mesh = meshSpec
		mockLocalMeshWorkloadClient.EXPECT().
			CreateMeshWorkload(ctx, discoveredMeshWorkload).
			Return(nil)

		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{*pod}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create MeshWorkload if Istio injected workload found later after Istio is discovered", func() {
		meshName := "meshName"
		meshNamespace := "meshNamespace"
		mesh := zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: meshName, Namespace: meshNamespace},
		}
		meshList := &zephyr_discovery.MeshList{Items: []zephyr_discovery.Mesh{mesh}}
		meshSpec := &zephyr_core_types.ResourceRef{
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
			GetMeshWorkload(ctx, objKey).
			Return(nil, notFoundErr)
		mockLocalMeshClient.EXPECT().
			ListMesh(ctx, &client.ListOptions{}).
			Return(meshList, nil)
		discoveredMeshWorkload.Spec.Mesh = meshSpec
		mockLocalMeshWorkloadClient.EXPECT().
			CreateMeshWorkload(ctx, discoveredMeshWorkload).
			Return(nil)

		// no pods when Istio is first discovered
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{}}, nil)

		// Now Istio has been discovered, but no pods will be found yet
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
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
		podCopy := *pod
		podCopy.ClusterName = clusterName
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, &podCopy).Return(nil, k8s_meta_types.ObjectMeta{}, expectedErr)
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{*pod}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})

		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
	})

	It("should return error if fatal error while populating Mesh resource ref", func() {
		expectedErr := eris.New("error")
		pod := &k8s_core_types.Pod{}
		podCopy := *pod
		podCopy.ClusterName = clusterName
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, &podCopy).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		mockLocalMeshClient.EXPECT().ListMesh(ctx, &client.ListOptions{}).Return(nil, expectedErr)
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{*pod}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})
		Expect(err).To(HaveOccurred())
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
	})

	It("should create new MeshWorkload if Istio injected workload updated", func() {
		newDiscoveredMeshWorkload := &zephyr_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: zephyr_discovery_types.MeshWorkloadSpec{
				KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &zephyr_core_types.ResourceRef{
						Namespace: "controller-namespace",
					},
				},
			},
		}
		newPod := &k8s_core_types.Pod{}
		mesh := zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &zephyr_discovery.MeshList{Items: []zephyr_discovery.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, newPod).
			Return(newDiscoveredMeshWorkload.Spec.KubeController.KubeControllerRef, newDiscoveredMeshWorkload.ObjectMeta, nil)
		mockLocalMeshClient.EXPECT().ListMesh(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
		Expect(pod.ClusterName).To(Equal(clusterName))
		Expect(newPod.ClusterName).To(Equal(clusterName))
	})

	It("should do nothing if MeshWorkload unchanged", func() {
		newDiscoveredMeshWorkload := &zephyr_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: zephyr_discovery_types.MeshWorkloadSpec{
				KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &zephyr_core_types.ResourceRef{
						Namespace: "controller-namespace",
					},
				},
			},
		}
		newPod := &k8s_core_types.Pod{}
		mesh := zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshList := &zephyr_discovery.MeshList{Items: []zephyr_discovery.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(discoveredMeshWorkload.Spec.KubeController.KubeControllerRef, discoveredMeshWorkload.ObjectMeta, nil)
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, newPod).
			Return(newDiscoveredMeshWorkload.Spec.KubeController.KubeControllerRef, newDiscoveredMeshWorkload.ObjectMeta, nil)
		mockLocalMeshClient.EXPECT().ListMesh(ctx, &client.ListOptions{}).Return(meshList, nil).Times(2)
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should update MeshWorkload if changed", func() {
		newNamespace := "new-controller-namespace"
		newDiscoveredMeshWorkload := &zephyr_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: zephyr_discovery_types.MeshWorkloadSpec{
				KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &zephyr_core_types.ResourceRef{
						Namespace: newNamespace,
					},
				},
			},
		}
		newPod := &k8s_core_types.Pod{}
		mesh := zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshSpec := &zephyr_core_types.ResourceRef{
			Name:      mesh.Name,
			Namespace: mesh.Namespace,
			Cluster:   clusterName,
		}
		meshList := &zephyr_discovery.MeshList{Items: []zephyr_discovery.Mesh{mesh}}
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
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return error if error processing old pod", func() {
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, k8s_meta_types.ObjectMeta{}, expectedErr)
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, &k8s_core_types.Pod{})
		Expect(err).To(HaveOccurred())
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
	})

	It("should return error if error processing new pod", func() {
		newPod := &k8s_core_types.Pod{}
		expectedErr := eris.New("error")
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, k8s_meta_types.ObjectMeta{}, nil)
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, newPod).Return(nil, k8s_meta_types.ObjectMeta{}, expectedErr)
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).To(HaveOccurred())
		Expect(err).To(testutils.HaveInErrorChain(expectedErr))
		Expect(testLogger.Sink().String()).To(ContainSubstring(mesh_workload.MeshWorkloadProcessingError))
	})

	It("should return nil if old and new Pods are not mesh injected", func() {
		mockMeshWorkloadScanner.EXPECT().ScanPod(ctx, pod).Return(nil, k8s_meta_types.ObjectMeta{}, nil).Times(2)
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, &k8s_core_types.Pod{})
		Expect(err).ToNot(HaveOccurred())
	})

	It("should create new MeshWorkload if pod is now Istio injected", func() {
		newPod := &k8s_core_types.Pod{}
		mesh := zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				Cluster: &zephyr_core_types.ResourceRef{Name: clusterName},
			},
			ObjectMeta: k8s_meta_types.ObjectMeta{Name: "meshName", Namespace: "meshNamespace"},
		}
		meshSpec := &zephyr_core_types.ResourceRef{
			Name:      mesh.Name,
			Namespace: mesh.Namespace,
			Cluster:   clusterName,
		}
		meshList := &zephyr_discovery.MeshList{Items: []zephyr_discovery.Mesh{mesh}}
		mockMeshWorkloadScanner.EXPECT().
			ScanPod(ctx, pod).
			Return(nil, k8s_meta_types.ObjectMeta{}, nil)
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
		podClient.EXPECT().
			ListPod(ctx).
			Return(&k8s_core_types.PodList{Items: []k8s_core_types.Pod{}}, nil)

		// Now Istio has been discovered
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: clusterName,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})

	It("does not start discovering if a mesh was discovered on some other cluster", func() {
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: "some-other-cluster-name",
				},
			},
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("does not remember that a mesh was discovered on some other cluster when handling a pod event later", func() {
		err := meshEventHandlerFuncs.OnCreate(&zephyr_discovery.Mesh{
			Spec: zephyr_discovery_types.MeshSpec{
				MeshType: &zephyr_discovery_types.MeshSpec_Istio{},
				Cluster: &zephyr_core_types.ResourceRef{
					Name: "some-other-cluster-name",
				},
			},
		})
		Expect(err).ToNot(HaveOccurred())

		newPod := &k8s_core_types.Pod{}

		err = podEventHandlerFuncs.OnUpdate(pod, newPod)
		Expect(err).ToNot(HaveOccurred())
	})
})
