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
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	"github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-service/k8s"
	discovery_mocks "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_corev1 "github.com/solo-io/service-mesh-hub/test/mocks/corev1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	k8s_core_types "k8s.io/api/core/v1"
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
			env.GetWriteNamespace(),
			serviceClient,
			meshServiceClient,
			meshWorkloadClient,
			meshClient,
		)
		return mocks{
			serviceClient:              serviceClient,
			meshServiceClient:          meshServiceClient,
			serviceEventWatcher:        serviceEventWatcher,
			meshWorkloadEventWatcher:   meshWorkloadEventWatcher,
			meshWorkloadClient:         meshWorkloadClient,
			meshClient:                 meshClient,
			meshServiceFinder:          meshServiceFinder,
			serviceCreateCallback:      serviceCreateCallback,
			serviceDeleteCallback:      serviceDeleteCallback,
			meshWorkloadCreateCallback: meshWorkloadCreateCallback,
			meshWorkloadDeleteCallback: meshWorkloadDeleteCallback,
		}
	}

	var expectReconcile = func(mocks mocks) {
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
				Namespace: env.GetWriteNamespace(),
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
				Namespace: env.GetWriteNamespace(),
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
		meshServiceToBeDeleted := &zephyr_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "service-with-no-corresponding-k8s-service",
				Namespace: "ns1",
			},
		}
		meshServiceName := "right-service-ns1-test-cluster-name"
		// this list call is the real one we care about
		mocks.meshWorkloadClient.EXPECT().
			ListMeshWorkload(ctx, client.MatchingLabels{
				constants.COMPUTE_TARGET: clusterName,
			}).
			Return(&zephyr_discovery.MeshWorkloadList{Items: []zephyr_discovery.MeshWorkload{*meshWorkloadEvent, *meshWorkloadEventV2}}, nil).
			Times(2)
		mocks.meshServiceClient.EXPECT().
			ListMeshService(ctx, client.MatchingLabels{
				constants.COMPUTE_TARGET: clusterName,
			}).
			Return(&zephyr_discovery.MeshServiceList{
				Items: []zephyr_discovery.MeshService{*meshServiceToBeDeleted},
			}, nil)
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
			Times(4)
		mocks.meshServiceClient.
			EXPECT().
			UpsertMeshServiceSpec(ctx, &zephyr_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      meshServiceName,
					Namespace: env.GetWriteNamespace(),
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
		mocks.meshServiceClient.
			EXPECT().
			DeleteMeshService(ctx, client.ObjectKey{
				Name:      meshServiceToBeDeleted.GetName(),
				Namespace: meshServiceToBeDeleted.GetNamespace(),
			}).
			Return(nil)
	}

	Context("mesh workload event", func() {
		It("reconciles MeshServices upon MeshWorkload creation", func() {
			mocks := setupMocks()
			expectReconcile(mocks)
			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")
			expectReconcile(mocks)
			err = (*mocks.meshWorkloadCreateCallback)(&zephyr_discovery.MeshWorkload{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("reconciles MeshServices upon MeshWorkload deletion", func() {
			mocks := setupMocks()
			expectReconcile(mocks)
			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")
			expectReconcile(mocks)
			err = (*mocks.meshWorkloadDeleteCallback)(&zephyr_discovery.MeshWorkload{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("reconciles MeshServices upon k8s Service creation", func() {
			mocks := setupMocks()
			expectReconcile(mocks)
			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")
			expectReconcile(mocks)
			err = (*mocks.serviceCreateCallback)(&k8s_core_types.Service{})
			Expect(err).NotTo(HaveOccurred())
		})
		It("reconciles MeshServices upon k8s Service deletion", func() {
			mocks := setupMocks()
			expectReconcile(mocks)
			err := mocks.meshServiceFinder.StartDiscovery(
				mocks.serviceEventWatcher,
				mocks.meshWorkloadEventWatcher,
			)
			Expect(err).NotTo(HaveOccurred(), "Should be able to start discovery")
			expectReconcile(mocks)
			err = (*mocks.serviceDeleteCallback)(&k8s_core_types.Service{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
