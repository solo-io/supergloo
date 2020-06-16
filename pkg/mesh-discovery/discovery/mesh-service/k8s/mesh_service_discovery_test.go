package k8s_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mock_kubernetes_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discovery_mocks "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-service/k8s"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mocks struct {
	serviceClient        *mock_kubernetes_core.MockServiceClient
	meshServiceClient    *discovery_mocks.MockMeshServiceClient
	meshWorkloadClient   *discovery_mocks.MockMeshWorkloadClient
	meshClient           *discovery_mocks.MockMeshClient
	meshServiceDiscovery k8s.MeshServiceDiscovery
}

var _ = Describe("Mesh Service Discovery", func() {
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
		meshServiceFinder := k8s.NewMeshServiceDiscovery(
			serviceClient,
			meshServiceClient,
			meshWorkloadClient,
			meshClient,
		)
		return mocks{
			serviceClient:        serviceClient,
			meshServiceClient:    meshServiceClient,
			meshWorkloadClient:   meshWorkloadClient,
			meshClient:           meshClient,
			meshServiceDiscovery: meshServiceFinder,
		}
	}

	var expectReconcile = func(mocks mocks) {
		workloadNamespace := "workload-namespace"
		mesh := &smh_discovery.Mesh{
			ObjectMeta: k8s_meta_types.ObjectMeta{
				Name:      "istio-test-mesh",
				Namespace: "istio-system",
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
			UpsertMeshService(ctx, &smh_discovery.MeshService{
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

	Context("MeshService discovery", func() {
		It("reconciles MeshServices", func() {
			mocks := setupMocks()
			expectReconcile(mocks)
			err := mocks.meshServiceDiscovery.DiscoverMeshServices(ctx, clusterName)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
