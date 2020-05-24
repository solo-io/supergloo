package k8s_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-service/k8s"
	discovery_mocks "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_corev1 "github.com/solo-io/service-mesh-hub/test/mocks/corev1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mocks struct {
	serviceClient            *mock_kubernetes_core.MockServiceClient
	meshServiceClient        *discovery_mocks.MockMeshServiceClient
	meshWorkloadClient       *discovery_mocks.MockMeshWorkloadClient
	meshClient               *discovery_mocks.MockMeshClient
	serviceEventWatcher      *mock_corev1.MockServiceEventWatcher
	meshWorkloadEventWatcher *mock_zephyr_discovery.MockMeshWorkloadEventWatcher

	meshServiceFinder k8s.MeshServiceFinder

	serviceCreateCallback      *func(service *k8s_core_types.Service) error
	serviceDeleteCallback      *func(service *k8s_core_types.Service) error
	meshWorkloadCreateCallback *func(meshWorkload *zephyr_discovery.MeshWorkload) error
	meshWorkloadDeleteCallback *func(meshWorkload *zephyr_discovery.MeshWorkload) error
}

var _ = Describe("Mesh Service Finder", func() {
	var (
		ctrl        *gomock.Controller
		ctx         = context.TODO()
		clusterName = "test-cluster-name"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var setupMocks = func() mocks {
		serviceClient := mock_kubernetes_core.NewMockServiceClient(ctrl)
		meshServiceClient := discovery_mocks.NewMockMeshServiceClient(ctrl)
		meshWorkloadClient := discovery_mocks.NewMockMeshWorkloadClient(ctrl)
		meshClient := discovery_mocks.NewMockMeshClient(ctrl)
		serviceEventWatcher := mock_corev1.NewMockServiceEventWatcher(ctrl)
		meshWorkloadEventWatcher := mock_zephyr_discovery.NewMockMeshWorkloadEventWatcher(ctrl)

		serviceCreateCallback := new(func(service *k8s_core_types.Service) error)
		meshWorkloadCreateCallback := new(func(meshWorkload *zephyr_discovery.MeshWorkload) error)

		serviceDeleteCallback := new(func(service *k8s_core_types.Service) error)
		meshWorkloadDeleteCallback := new(func(meshWorkload *zephyr_discovery.MeshWorkload) error)

		// need to grab the callbacks so we can hook into them and send events
		serviceEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, serviceEventHandler *k8s_core_controller.ServiceEventHandlerFuncs) error {
				*serviceCreateCallback = serviceEventHandler.OnCreate
				*serviceDeleteCallback = serviceEventHandler.OnDelete
				return nil
			})

		meshWorkloadEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, mwEventHandler *controller.MeshWorkloadEventHandlerFuncs) error {
				*meshWorkloadCreateCallback = mwEventHandler.OnCreate
				*meshWorkloadDeleteCallback = mwEventHandler.OnDelete
				return nil
			})

		meshServiceFinder := k8s.NewMeshServiceFinder(
			ctx,
			clusterName,
			container_runtime.GetWriteNamespace(),
			serviceClient,
			meshServiceClient,
			meshWorkloadClient,
			meshClient,
		)

		return mocks{
			serviceClient:            serviceClient,
			meshServiceClient:        meshServiceClient,
			serviceEventWatcher:      serviceEventWatcher,
			meshWorkloadEventWatcher: meshWorkloadEventWatcher,
			meshWorkloadClient:       meshWorkloadClient,
			meshClient:               meshClient,

			meshServiceFinder: meshServiceFinder,

			serviceCreateCallback:      serviceCreateCallback,
			serviceDeleteCallback:      serviceDeleteCallback,
			meshWorkloadCreateCallback: meshWorkloadCreateCallback,
			meshWorkloadDeleteCallback: meshWorkloadDeleteCallback,
		}
	}

	Context("mesh workload event", func() {
		It("can associate a mesh workload with an existing service", func() {
			mocks := setupMocks()

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}

			meshWorkloadEvent := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v1",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			meshWorkloadEventV2 := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload-v2",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v2",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			wrongService := k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "wrong-service",
					Namespace: "ns1",
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"other-label": "value",
					},
					Ports: []k8s_core_types.ServicePort{{
						Name:       "port-1",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.IntOrString{IntVal: 8080},
						NodePort:   32000,
					}},
				},
			}
			rightService := k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "right-service",
					Namespace: "ns1",
					Labels: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"label": "value",
					},
					Ports: []k8s_core_types.ServicePort{{
						Name:       "correct-service-port",
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.IntOrString{IntVal: 8443},
						NodePort:   32001,
					}},
				},
			}

			meshServiceName := "right-service-ns1-test-cluster-name"

			// first list call is on startup, making it empty just for convenience in this test
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)

			// this list call is the real one we care about
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{Items: []zephyr_discovery.MeshWorkload{*meshWorkloadEvent, *meshWorkloadEventV2}}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)
			mocks.serviceClient.
				EXPECT().
				ListService(ctx).
				Return(&k8s_core_types.ServiceList{
					Items: []k8s_core_types.Service{wrongService, rightService},
				}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(mesh, nil).
				Times(3)

			mocks.meshServiceClient.
				EXPECT().
				GetMeshService(ctx, client.ObjectKey{
					Name:      meshServiceName,
					Namespace: container_runtime.GetWriteNamespace(),
				}).
				Return(nil, errors.NewNotFound(zephyr_discovery.Resource("meshservice"), meshServiceName))

			mocks.meshServiceClient.
				EXPECT().
				CreateMeshService(ctx, &zephyr_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      meshServiceName,
						Namespace: container_runtime.GetWriteNamespace(),
						Labels:    k8s.DiscoveryLabels(zephyr_core_types.MeshType_LINKERD, clusterName, rightService.GetName(), rightService.GetNamespace()),
					},
					Spec: zephyr_discovery_types.MeshServiceSpec{
						KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
							Ref: &zephyr_core_types.ResourceRef{
								Name:      rightService.GetName(),
								Namespace: rightService.GetNamespace(),
								Cluster:   clusterName,
							},
							WorkloadSelectorLabels: rightService.Spec.Selector,
							Labels:                 rightService.GetLabels(),
							Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Name:     "correct-service-port",
								Port:     443,
								Protocol: "TCP",
							}},
						},
						Mesh: meshWorkloadEvent.Spec.Mesh,
						Subsets: map[string]*zephyr_discovery_types.MeshServiceSpec_Subset{
							"version": {
								Values: []string{"v1", "v2"},
							},
						},
					},
				}).
				Return(nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.meshWorkloadCreateCallback)(meshWorkloadEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not associate a mesh workload to any service if the labels don't match", func() {
			mocks := setupMocks()

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}

			meshWorkloadEvent := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			wrongService := k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "wrong-service",
					Namespace: "ns1",
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"other-label": "value",
					},
				},
			}

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)

			mocks.serviceClient.
				EXPECT().
				ListService(ctx).
				Return(&k8s_core_types.ServiceList{
					Items: []k8s_core_types.Service{wrongService, wrongService},
				}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(mesh, nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.meshWorkloadCreateCallback)(meshWorkloadEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("bails out early if the mesh workload has no labels to match on", func() {
			mocks := setupMocks()

			meshWorkloadEvent := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: nil,
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.meshWorkloadCreateCallback)(meshWorkloadEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not match a service with no labels to the mesh workload event", func() {
			mocks := setupMocks()

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}

			meshWorkloadEvent := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			wrongService := k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "wrong-service",
					Namespace: "ns1",
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{},
				},
			}

			mocks.serviceClient.
				EXPECT().
				ListService(ctx).
				Return(&k8s_core_types.ServiceList{
					Items: []k8s_core_types.Service{wrongService, wrongService},
				}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(mesh, nil)

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.meshWorkloadCreateCallback)(meshWorkloadEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not create a mesh service if it already exists", func() {
			mocks := setupMocks()

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}

			meshWorkloadEvent := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			wrongService := k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "wrong-service",
					Namespace: "ns1",
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"other-label": "value",
					},
				},
			}
			rightService := k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "right-service",
					Namespace: "ns1",
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"label": "value",
					},
					Ports: []k8s_core_types.ServicePort{{
						Name:       "port-1",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.IntOrString{IntVal: 8080},
						NodePort:   32000,
					}},
				},
			}

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)
			mocks.serviceClient.
				EXPECT().
				ListService(ctx).
				Return(&k8s_core_types.ServiceList{
					Items: []k8s_core_types.Service{wrongService, rightService},
				}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(mesh, nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.meshWorkloadCreateCallback)(meshWorkloadEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can delete the appropriate mesh service when a workload is deleted", func() {
			mocks := setupMocks()
			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}
			meshWorkloadEvent := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v1",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			meshWorkloadEventV2 := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload-v2",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "wont-get-selected-by-service",
							"version":              "v2",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			kubeService := &k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "right-service",
					Namespace: "ns1",
					Labels: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"label": "value",
					},
					Ports: []k8s_core_types.ServicePort{{
						Name:       "correct-service-port",
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.IntOrString{IntVal: 8443},
						NodePort:   32001,
					}},
				},
			}
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: clients.ObjectMetaToResourceRef(kubeService.ObjectMeta),
					},
				},
			}

			// first list call is on startup, making it empty just for convenience in this test
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{Items: []zephyr_discovery.MeshWorkload{*meshWorkloadEvent, *meshWorkloadEventV2}}, nil)
			mocks.serviceClient.EXPECT().
				GetService(ctx, clients.ResourceRefToObjectKey(meshService.Spec.KubeService.Ref)).
				Return(kubeService, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{Items: []zephyr_discovery.MeshService{*meshService}}, nil)
			mocks.meshServiceClient.EXPECT().
				DeleteMeshService(ctx, clients.ObjectMetaToObjectKey(meshService.ObjectMeta)).
				Return(nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.meshWorkloadDeleteCallback)(meshWorkloadEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can delete the appropriate mesh service if its kube service has already been deleted when the event arrives", func() {
			mocks := setupMocks()
			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}
			meshWorkloadEvent := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v1",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			meshWorkloadEventV2 := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload-v2",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "wont-get-selected-by-service",
							"version":              "v2",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			kubeService := &k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "right-service",
					Namespace: "ns1",
					Labels: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"label": "value",
					},
					Ports: []k8s_core_types.ServicePort{{
						Name:       "correct-service-port",
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.IntOrString{IntVal: 8443},
						NodePort:   32001,
					}},
				},
			}
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: clients.ObjectMetaToResourceRef(kubeService.ObjectMeta),
					},
				},
			}

			// first list call is on startup, making it empty just for convenience in this test
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{Items: []zephyr_discovery.MeshWorkload{*meshWorkloadEvent, *meshWorkloadEventV2}}, nil)
			mocks.serviceClient.EXPECT().
				GetService(ctx, clients.ResourceRefToObjectKey(meshService.Spec.KubeService.Ref)).
				Return(nil, errors.NewNotFound(zephyr_discovery.Resource("kube service"), ""))
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{Items: []zephyr_discovery.MeshService{*meshService}}, nil)
			mocks.meshServiceClient.EXPECT().
				DeleteMeshService(ctx, clients.ObjectMetaToObjectKey(meshService.ObjectMeta)).
				Return(nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.meshWorkloadDeleteCallback)(meshWorkloadEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not delete a service if there are still other backing workloads", func() {
			mocks := setupMocks()
			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}
			meshWorkloadEvent := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v1",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			meshWorkloadEventV2 := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload-v2",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v2",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			kubeService := &k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "right-service",
					Namespace: "ns1",
					Labels: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"label": "value",
					},
					Ports: []k8s_core_types.ServicePort{{
						Name:       "correct-service-port",
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.IntOrString{IntVal: 8443},
						NodePort:   32001,
					}},
				},
			}
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: clients.ObjectMetaToResourceRef(kubeService.ObjectMeta),
					},
				},
			}

			// first list call is on startup, making it empty just for convenience in this test
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{Items: []zephyr_discovery.MeshWorkload{*meshWorkloadEvent, *meshWorkloadEventV2}}, nil)
			mocks.serviceClient.EXPECT().
				GetService(ctx, clients.ResourceRefToObjectKey(meshService.Spec.KubeService.Ref)).
				Return(kubeService, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{Items: []zephyr_discovery.MeshService{*meshService}}, nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.meshWorkloadDeleteCallback)(meshWorkloadEvent)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("service event", func() {
		It("can associate a service with an existing mesh workload", func() {
			mocks := setupMocks()

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}

			serviceEvent := &k8s_core_types.Service{
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"app":   "test-app",
						"track": "canary",
					},
					Ports: []k8s_core_types.ServicePort{{
						Name:       "port-1",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.IntOrString{IntVal: 8080},
						NodePort:   32000,
					}},
				},
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "my-svc",
					Namespace: "my-ns",
					Labels: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
			}

			wrongWorkload := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":   "test-app",
							"track": "production",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}
			rightWorkloadV1 := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":     "test-app",
							"track":   "canary",
							"version": "v1",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}
			rightWorkloadV2 := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":     "test-app",
							"track":   "canary",
							"version": "v2",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}

			meshServiceName := "my-svc-my-ns-test-cluster-name"

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)
			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, client.ObjectKey{
					Name:      mesh.Name,
					Namespace: mesh.Namespace,
				}).
				Return(mesh, nil).
				Times(3)

			mocks.meshWorkloadClient.
				EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{
					Items: []zephyr_discovery.MeshWorkload{
						*wrongWorkload,
						*rightWorkloadV1,
						*rightWorkloadV2,
					},
				}, nil)

			mocks.meshServiceClient.
				EXPECT().
				GetMeshService(ctx, client.ObjectKey{
					Name:      meshServiceName,
					Namespace: container_runtime.GetWriteNamespace(),
				}).
				Return(nil, errors.NewNotFound(zephyr_discovery.Resource("meshservice"), meshServiceName))

			mocks.meshServiceClient.
				EXPECT().
				CreateMeshService(ctx, &zephyr_discovery.MeshService{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      meshServiceName,
						Namespace: container_runtime.GetWriteNamespace(),
						Labels:    k8s.DiscoveryLabels(zephyr_core_types.MeshType_LINKERD, clusterName, serviceEvent.GetName(), serviceEvent.GetNamespace()),
					},
					Spec: zephyr_discovery_types.MeshServiceSpec{
						KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
							Ref: &zephyr_core_types.ResourceRef{
								Name:      serviceEvent.GetName(),
								Namespace: serviceEvent.GetNamespace(),
								Cluster:   clusterName,
							},
							WorkloadSelectorLabels: serviceEvent.Spec.Selector,
							Labels:                 serviceEvent.GetLabels(),
							Ports: []*zephyr_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Name:     "port-1",
								Port:     80,
								Protocol: "TCP",
							}},
						},
						Mesh: rightWorkloadV1.Spec.Mesh,
						Subsets: map[string]*zephyr_discovery_types.MeshServiceSpec_Subset{
							"version": {
								Values: []string{"v1", "v2"},
							},
						},
					},
				}).
				Return(nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.serviceCreateCallback)(serviceEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not associate a service to any mesh workload if the labels don't match", func() {
			mocks := setupMocks()

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}

			serviceEvent := &k8s_core_types.Service{
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"app":   "test-app",
						"track": "canary",
					},
				},
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "my-svc",
					Namespace: "my-ns",
				},
			}

			wrongWorkload1 := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":   "test-app",
							"track": "production",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			wrongWorkload2 := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":   "test-other-unrelated-app",
							"track": "canary",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)
			mocks.meshWorkloadClient.
				EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{
					Items: []zephyr_discovery.MeshWorkload{
						*wrongWorkload1,
						*wrongWorkload2,
					},
				}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, client.ObjectKey{
					Name:      mesh.Name,
					Namespace: mesh.Namespace,
				}).
				Return(mesh, nil).
				Times(2)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.serviceCreateCallback)(serviceEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("will bail out early if the mesh is on a different cluster than the service", func() {
			mocks := setupMocks()

			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: "incorrect-cluster-name",
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}

			serviceEvent := &k8s_core_types.Service{
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"app":   "test-app",
						"track": "canary",
					},
				},
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "my-svc",
					Namespace: "my-ns",
				},
			}

			wrongWorkload1 := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":   "test-app",
							"track": "production",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)
			mocks.meshWorkloadClient.
				EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{
					Items: []zephyr_discovery.MeshWorkload{
						*wrongWorkload1,
					},
				}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, client.ObjectKey{
					Name:      mesh.Name,
					Namespace: mesh.Namespace,
				}).
				Return(mesh, nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.serviceCreateCallback)(serviceEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("bails out early if the mesh workload has no labels to match on", func() {
			mocks := setupMocks()

			serviceEvent := &k8s_core_types.Service{
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{},
				},
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "my-svc",
					Namespace: "my-ns",
				},
			}

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.serviceCreateCallback)(serviceEvent)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can delete the appropriate mesh service when its kube service is deleted", func() {
			mocks := setupMocks()

			serviceEvent := &k8s_core_types.Service{
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"app":   "test-app",
						"track": "canary",
					},
					Ports: []k8s_core_types.ServicePort{{
						Name:       "port-1",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.IntOrString{IntVal: 8080},
						NodePort:   32000,
					}},
				},
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "my-svc",
					Namespace: "my-ns",
					Labels: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
			}
			wrongWorkload := &zephyr_discovery.MeshWorkload{
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":   "test-app",
							"track": "production",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}
			rightWorkloadV1 := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":     "test-app",
							"track":   "canary",
							"version": "v1",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}
			rightWorkloadV2 := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":     "test-app",
							"track":   "canary",
							"version": "v2",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}
			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{
					Items: []zephyr_discovery.MeshWorkload{
						*wrongWorkload,
						*rightWorkloadV1,
						*rightWorkloadV2,
					},
				}, nil)
			mocks.meshClient.EXPECT().
				GetMesh(ctx, clients.ResourceRefToObjectKey(wrongWorkload.Spec.Mesh)).
				Return(mesh, nil)
			mocks.meshClient.EXPECT().
				GetMesh(ctx, clients.ResourceRefToObjectKey(rightWorkloadV1.Spec.Mesh)).
				Return(mesh, nil)
			mocks.meshClient.EXPECT().
				GetMesh(ctx, clients.ResourceRefToObjectKey(rightWorkloadV2.Spec.Mesh)).
				Return(mesh, nil)
			mocks.meshServiceClient.EXPECT().
				DeleteMeshService(ctx, client.ObjectKey{
					Name:      "my-svc-my-ns-test-cluster-name",
					Namespace: container_runtime.GetWriteNamespace(),
				}).
				Return(nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")

			err = (*mocks.serviceDeleteCallback)(serviceEvent)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("garbage collection", func() {
		It("does nothing if there is nothing to do", func() {
			mocks := setupMocks()

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{}, nil)

			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{}, nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")
		})

		It("cleans up mesh services if their services have disappeared", func() {
			mocks := setupMocks()
			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}
			meshWorkload := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v1",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: &zephyr_core_types.ResourceRef{
							Name:      "doesn't-exist",
							Namespace: "whoops-deleted",
						},
					},
				},
			}

			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{Items: []zephyr_discovery.MeshWorkload{*meshWorkload}}, nil)
			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{Items: []zephyr_discovery.MeshService{*meshService}}, nil)
			mocks.serviceClient.EXPECT().
				GetService(ctx, clients.ResourceRefToObjectKey(meshService.Spec.KubeService.Ref)).
				Return(nil, errors.NewNotFound(zephyr_discovery.Resource("meshservice"), meshService.GetName()))
			mocks.meshServiceClient.EXPECT().
				DeleteMeshService(ctx, clients.ObjectMetaToObjectKey(meshService.ObjectMeta)).
				Return(nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")
		})

		It("does not delete a service if there are still other backing workloads", func() {
			mocks := setupMocks()
			mesh := &zephyr_discovery.Mesh{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: zephyr_discovery_types.MeshSpec{
					Cluster: &zephyr_core_types.ResourceRef{
						Name: clusterName,
					},
					MeshType: &zephyr_discovery_types.MeshSpec_Linkerd{},
				},
			}
			meshWorkloadEvent := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v1",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			meshWorkloadEventV2 := &zephyr_discovery.MeshWorkload{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "test-mesh-workload-v2",
					Namespace: container_runtime.GetWriteNamespace(),
					Labels: map[string]string{
						constants.COMPUTE_TARGET: clusterName,
					},
				},
				Spec: zephyr_discovery_types.MeshWorkloadSpec{
					KubeController: &zephyr_discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v2",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &zephyr_core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			kubeService := &k8s_core_types.Service{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      "right-service",
					Namespace: "ns1",
					Labels: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
				Spec: k8s_core_types.ServiceSpec{
					Selector: map[string]string{
						"label": "value",
					},
					Ports: []k8s_core_types.ServicePort{{
						Name:       "correct-service-port",
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.IntOrString{IntVal: 8443},
						NodePort:   32001,
					}},
				},
			}
			meshService := &zephyr_discovery.MeshService{
				Spec: zephyr_discovery_types.MeshServiceSpec{
					KubeService: &zephyr_discovery_types.MeshServiceSpec_KubeService{
						Ref: clients.ObjectMetaToResourceRef(kubeService.ObjectMeta),
					},
				},
			}

			mocks.meshServiceClient.EXPECT().
				ListMeshService(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshServiceList{Items: []zephyr_discovery.MeshService{*meshService}}, nil)
			mocks.meshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.COMPUTE_TARGET: clusterName,
				}).
				Return(&zephyr_discovery.MeshWorkloadList{Items: []zephyr_discovery.MeshWorkload{*meshWorkloadEvent, *meshWorkloadEventV2}}, nil)
			mocks.serviceClient.EXPECT().
				GetService(ctx, clients.ResourceRefToObjectKey(meshService.Spec.KubeService.Ref)).
				Return(kubeService, nil)

			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")
		})
	})
})
