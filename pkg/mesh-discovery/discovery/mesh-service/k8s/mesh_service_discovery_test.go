package k8s_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8s_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	mock_kubernetes_core "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/mocks"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	discovery_mocks "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/mocks"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/selection"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/discovery/mesh-service/k8s"
	mock_multicluster "github.com/solo-io/service-mesh-hub/test/mocks/smh/clients"
	k8s_core_types "k8s.io/api/core/v1"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Mesh Service Discovery", func() {
	var (
		ctrl                   *gomock.Controller
		ctx                    = context.TODO()
		clusterName            = "test-cluster-name"
		mockMeshServiceClient  *discovery_mocks.MockMeshServiceClient
		mockMeshWorkloadClient *discovery_mocks.MockMeshWorkloadClient
		mockMeshClient         *discovery_mocks.MockMeshClient
		mockMcClient           *mock_multicluster.MockClient
		mockServiceClient      *mock_kubernetes_core.MockServiceClient
		meshServiceDiscovery   k8s.MeshServiceDiscovery
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockMeshServiceClient = discovery_mocks.NewMockMeshServiceClient(ctrl)
		mockMeshWorkloadClient = discovery_mocks.NewMockMeshWorkloadClient(ctrl)
		mockMeshClient = discovery_mocks.NewMockMeshClient(ctrl)
		mockServiceClient = mock_kubernetes_core.NewMockServiceClient(ctrl)
		mockMcClient = mock_multicluster.NewMockClient(ctrl)
		meshServiceDiscovery = k8s.NewMeshServiceDiscovery(
			mockMeshServiceClient,
			mockMeshWorkloadClient,
			mockMeshClient,
			func(client client.Client) k8s_core.ServiceClient {
				return mockServiceClient
			},
			mockMcClient,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var expectReconcile = func() {
		// Doesn't matter what's returned because the mockServiceClientFactory will always return mockServiceClient
		mockMcClient.EXPECT().Cluster(clusterName).Return(nil, nil)
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
		mockMeshWorkloadClient.EXPECT().
			ListMeshWorkload(ctx, client.MatchingLabels{
				kube.COMPUTE_TARGET: clusterName,
			}).
			Return(&smh_discovery.MeshWorkloadList{Items: []smh_discovery.MeshWorkload{*meshWorkloadEvent, *meshWorkloadEventV2}}, nil).
			Times(2)
		mockMeshServiceClient.EXPECT().
			ListMeshService(ctx, client.MatchingLabels{
				kube.COMPUTE_TARGET: clusterName,
			}).
			Return(&smh_discovery.MeshServiceList{
				Items: []smh_discovery.MeshService{*meshServiceToBeDeleted},
			}, nil)
		mockServiceClient.
			EXPECT().
			ListService(ctx).
			Return(&k8s_core_types.ServiceList{
				Items: []k8s_core_types.Service{wrongService, rightService},
			}, nil)
		mockMeshClient.
			EXPECT().
			GetMesh(ctx, selection.ObjectMetaToObjectKey(mesh.ObjectMeta)).
			Return(mesh, nil).
			Times(4)
		mockMeshServiceClient.
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
		mockMeshServiceClient.
			EXPECT().
			DeleteMeshService(ctx, client.ObjectKey{
				Name:      meshServiceToBeDeleted.GetName(),
				Namespace: meshServiceToBeDeleted.GetNamespace(),
			}).
			Return(nil)
	}

	Context("MeshService discovery", func() {
		It("reconciles MeshServices", func() {
			expectReconcile()
			err := meshServiceDiscovery.DiscoverMeshServices(ctx, clusterName)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
