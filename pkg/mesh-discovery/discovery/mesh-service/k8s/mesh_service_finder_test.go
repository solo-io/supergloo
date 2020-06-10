package k8s_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/controller"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	k8s_core_controller "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-service/k8s"
	discovery_mocks "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.smh.solo.io/v1alpha1"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/core/v1"
	mock_corev1 "github.com/solo-io/service-mesh-hub/test/mocks/corev1"
	mock_smh_discovery "github.com/solo-io/service-mesh-hub/test/mocks/smh/discovery"
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
	meshWorkloadEventWatcher *mock_smh_discovery.MockMeshWorkloadEventWatcher

	meshServiceFinder k8s.MeshServiceFinder

	serviceCreateCallback      *func(service *k8s_core_types.Service) error
	serviceDeleteCallback      *func(service *k8s_core_types.Service) error
	meshWorkloadCreateCallback *func(meshWorkload *smh_discovery.MeshWorkload) error
	meshWorkloadDeleteCallback *func(meshWorkload *smh_discovery.MeshWorkload) error
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
		meshWorkloadEventWatcher := mock_smh_discovery.NewMockMeshWorkloadEventWatcher(ctrl)
		serviceCreateCallback := new(func(service *k8s_core_types.Service) error)
		meshWorkloadCreateCallback := new(func(meshWorkload *smh_discovery.MeshWorkload) error)
		serviceDeleteCallback := new(func(service *k8s_core_types.Service) error)
		meshWorkloadDeleteCallback := new(func(meshWorkload *smh_discovery.MeshWorkload) error)
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
		workloadNamespace := "workload-namespace"
		mesh := &smh_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "istio-test-mesh",
				Namespace: "isito-system",
			},
			Spec: smh_discovery_types.MeshSpec{
				Cluster: &smh_core_types.ResourceRef{
					Name: clusterName,
				},
				MeshType: &smh_discovery_types.MeshSpec_Linkerd{},
			},
		}
		meshWorkloadEvent := &smh_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "test-mesh-workload",
				Namespace: container_runtime.GetWriteNamespace(),
				Labels: map[string]string{
					kube.COMPUTE_TARGET: clusterName,
				},
			},
			Spec: smh_discovery_types.MeshWorkloadSpec{
				KubeController: &smh_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &smh_core_types.ResourceRef{
						Namespace: workloadNamespace,
					},
					Labels: map[string]string{
						"label":                "value",
						"version":              "v1",
						"istio-injected-label": "doesn't matter",
					},
				},
				Mesh: &smh_core_types.ResourceRef{
					Name:      mesh.Name,
					Namespace: mesh.Namespace,
				},
			},
		}
		meshWorkloadEventV2 := &smh_discovery.MeshWorkload{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "test-mesh-workload-v2",
				Namespace: container_runtime.GetWriteNamespace(),
				Labels: map[string]string{
					kube.COMPUTE_TARGET: clusterName,
				},
			},
			Spec: smh_discovery_types.MeshWorkloadSpec{
				KubeController: &smh_discovery_types.MeshWorkloadSpec_KubeController{
					KubeControllerRef: &smh_core_types.ResourceRef{
						Namespace: workloadNamespace,
					},
					Labels: map[string]string{
						"label":                "value",
						"version":              "v2",
						"istio-injected-label": "doesn't matter",
					},
				},
				Mesh: &smh_core_types.ResourceRef{
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
				Namespace: workloadNamespace,
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
		meshServiceToBeDeleted := &smh_discovery.MeshService{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "service-with-no-corresponding-k8s-service",
				Namespace: "ns1",
			},
		}
		meshServiceName := fmt.Sprintf("right-service-%s-test-cluster-name", workloadNamespace)
		// this list call is the real one we care about
		mocks.meshWorkloadClient.EXPECT().
			ListMeshWorkload(ctx, client.MatchingLabels{
				kube.COMPUTE_TARGET: clusterName,
			}).
			Return(&smh_discovery.MeshWorkloadList{Items: []smh_discovery.MeshWorkload{*meshWorkloadEvent, *meshWorkloadEventV2}}, nil).
			Times(2)
		mocks.meshServiceClient.EXPECT().
			ListMeshService(ctx, client.MatchingLabels{
				kube.COMPUTE_TARGET: clusterName,
			}).
			Return(&smh_discovery.MeshServiceList{
				Items: []smh_discovery.MeshService{*meshServiceToBeDeleted},
			}, nil)
		mocks.serviceClient.
			EXPECT().
			ListService(ctx).
			Return(&k8s_core_types.ServiceList{
				Items: []k8s_core_types.Service{wrongService, rightService},
			}, nil)
		mocks.meshClient.
			EXPECT().
			GetMesh(ctx, selection.ObjectMetaToObjectKey(mesh.ObjectMeta)).
			Return(mesh, nil).
			Times(4)
		mocks.meshServiceClient.
			EXPECT().
			UpsertMeshServiceSpec(ctx, &smh_discovery.MeshService{
				ObjectMeta: k8s_meta_types.ObjectMeta{
					Name:      meshServiceName,
					Namespace: container_runtime.GetWriteNamespace(),
					Labels:    k8s.DiscoveryLabels(smh_core_types.MeshType_LINKERD, clusterName, rightService.GetName(), rightService.GetNamespace()),
				},
				Spec: smh_discovery_types.MeshServiceSpec{
					KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
						Ref: &smh_core_types.ResourceRef{
							Name:      rightService.GetName(),
							Namespace: rightService.GetNamespace(),
							Cluster:   clusterName,
						},
						WorkloadSelectorLabels: rightService.Spec.Selector,
						Labels:                 rightService.GetLabels(),
						Ports: []*smh_discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
							Name:     "correct-service-port",
							Port:     443,
							Protocol: "TCP",
						}},
					},
					Mesh: meshWorkloadEvent.Spec.Mesh,
					Subsets: map[string]*smh_discovery_types.MeshServiceSpec_Subset{
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
			err = (*mocks.meshWorkloadCreateCallback)(&smh_discovery.MeshWorkload{})
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
			err = (*mocks.meshWorkloadDeleteCallback)(&smh_discovery.MeshWorkload{})
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
