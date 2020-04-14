package mesh_service_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mesh_service "github.com/solo-io/service-mesh-hub/services/mesh-discovery/pkg/discovery/mesh-service"
	discovery_mocks "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_corev1 "github.com/solo-io/service-mesh-hub/test/mocks/corev1"
	mock_zephyr_discovery "github.com/solo-io/service-mesh-hub/test/mocks/zephyr/discovery"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	meshServiceFinder mesh_service.MeshServiceFinder

	serviceCallback      func(service *corev1.Service) error
	meshWorkloadCallback func(meshWorkload *v1alpha1.MeshWorkload) error
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
		meshWorkloadController := mock_zephyr_discovery.NewMockMeshWorkloadEventWatcher(ctrl)

		var serviceCallback func(service *corev1.Service) error
		var meshWorkloadCallback func(meshWorkload *v1alpha1.MeshWorkload) error

		// need to grab the callbacks so we can hook into them and send events
		serviceEventWatcher.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, serviceEventHandler *mesh_service.ServiceEventHandler) error {
				serviceCallback = serviceEventHandler.HandleServiceUpsert
				return nil
			})

		meshWorkloadController.
			EXPECT().
			AddEventHandler(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, mwEventHandler *mesh_service.MeshWorkloadEventHandler) error {
				meshWorkloadCallback = mwEventHandler.HandleMeshWorkloadUpsert
				return nil
			})

		meshServiceFinder := mesh_service.NewMeshServiceFinder(
			ctx,
			clusterName,
			env.GetWriteNamespace(),
			serviceClient,
			meshServiceClient,
			meshWorkloadClient,
			meshClient,
		)

		err := meshServiceFinder.StartDiscovery(
			serviceEventWatcher,
			meshWorkloadController,
		)
		Expect(err).NotTo(HaveOccurred())

		return mocks{
			serviceClient:            serviceClient,
			meshServiceClient:        meshServiceClient,
			serviceEventWatcher:      serviceEventWatcher,
			meshWorkloadEventWatcher: meshWorkloadController,
			meshWorkloadClient:       meshWorkloadClient,
			meshClient:               meshClient,

			meshServiceFinder: meshServiceFinder,

			serviceCallback:      serviceCallback,
			meshWorkloadCallback: meshWorkloadCallback,
		}
	}

	Context("mesh workload event", func() {
		It("can associate a mesh workload with an existing service", func() {
			mocks := setupMocks()

			mesh := &v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
				},
			}

			meshWorkloadEvent := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v1",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			meshWorkloadEventV2 := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"version":              "v2",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			wrongService := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong-service",
					Namespace: "ns1",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"other-label": "value",
					},
					Ports: []corev1.ServicePort{{
						Name:       "port-1",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.IntOrString{IntVal: 8080},
						NodePort:   32000,
					}},
				},
			}
			rightService := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "right-service",
					Namespace: "ns1",
					Labels: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"label": "value",
					},
					Ports: []corev1.ServicePort{{
						Name:       "correct-service-port",
						Protocol:   "TCP",
						Port:       443,
						TargetPort: intstr.IntOrString{IntVal: 8443},
						NodePort:   32001,
					}},
				},
			}

			meshServiceName := "right-service-ns1-test-cluster-name"

			mocks.serviceClient.
				EXPECT().
				List(ctx).
				Return(&corev1.ServiceList{
					Items: []corev1.Service{wrongService, rightService},
				}, nil)

			mocks.meshWorkloadClient.
				EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{Items: []v1alpha1.MeshWorkload{
					*meshWorkloadEvent,
					*meshWorkloadEventV2,
				}}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(mesh, nil)

			mocks.meshServiceClient.
				EXPECT().
				GetMeshService(ctx, client.ObjectKey{
					Name:      meshServiceName,
					Namespace: env.GetWriteNamespace(),
				}).
				Return(nil, errors.NewNotFound(v1alpha1.Resource("meshservice"), meshServiceName))

			mocks.meshServiceClient.
				EXPECT().
				CreateMeshService(ctx, &v1alpha1.MeshService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      meshServiceName,
						Namespace: env.GetWriteNamespace(),
						Labels:    mesh_service.DiscoveryLabels(clusterName, rightService.GetName(), rightService.GetNamespace()),
					},
					Spec: discovery_types.MeshServiceSpec{
						KubeService: &discovery_types.MeshServiceSpec_KubeService{
							Ref: &core_types.ResourceRef{
								Name:      rightService.GetName(),
								Namespace: rightService.GetNamespace(),
								Cluster:   clusterName,
							},
							WorkloadSelectorLabels: rightService.Spec.Selector,
							Labels:                 rightService.GetLabels(),
							Ports: []*discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Name:     "correct-service-port",
								Port:     443,
								Protocol: "TCP",
							}},
						},
						Mesh: meshWorkloadEvent.Spec.Mesh,
						Subsets: map[string]*discovery_types.MeshServiceSpec_Subset{
							"version": {
								Values: []string{"v1", "v2"},
							},
						},
					},
				}).
				Return(nil)

			err := mocks.meshWorkloadCallback(meshWorkloadEvent)

			Expect(err).NotTo(HaveOccurred())
		})

		It("does not associate a mesh workload to any service if the labels don't match", func() {
			mocks := setupMocks()

			mesh := &v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
				},
			}

			meshWorkloadEvent := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			wrongService := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong-service",
					Namespace: "ns1",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"other-label": "value",
					},
				},
			}

			mocks.serviceClient.
				EXPECT().
				List(ctx).
				Return(&corev1.ServiceList{
					Items: []corev1.Service{wrongService, wrongService},
				}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(mesh, nil)

			err := mocks.meshWorkloadCallback(meshWorkloadEvent)

			Expect(err).NotTo(HaveOccurred())
		})

		It("bails out early if the mesh workload has no labels to match on", func() {
			mocks := setupMocks()

			meshWorkloadEvent := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: nil,
					},
					Mesh: &core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}

			err := mocks.meshWorkloadCallback(meshWorkloadEvent)

			Expect(err).NotTo(HaveOccurred())
		})

		It("does not match a service with no labels to the mesh workload event", func() {
			mocks := setupMocks()

			mesh := &v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
				},
			}

			meshWorkloadEvent := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			wrongService := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong-service",
					Namespace: "ns1",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{},
				},
			}

			mocks.serviceClient.
				EXPECT().
				List(ctx).
				Return(&corev1.ServiceList{
					Items: []corev1.Service{wrongService, wrongService},
				}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(mesh, nil)

			err := mocks.meshWorkloadCallback(meshWorkloadEvent)

			Expect(err).NotTo(HaveOccurred())
		})

		It("does not create a mesh service if it already exists", func() {
			mocks := setupMocks()

			mesh := &v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
				},
			}

			meshWorkloadEvent := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"label":                "value",
							"istio-injected-label": "doesn't matter",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			wrongService := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wrong-service",
					Namespace: "ns1",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"other-label": "value",
					},
				},
			}
			rightService := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "right-service",
					Namespace: "ns1",
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"label": "value",
					},
					Ports: []corev1.ServicePort{{
						Name:       "port-1",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.IntOrString{IntVal: 8080},
						NodePort:   32000,
					}},
				},
			}

			meshServiceName := "right-service-ns1-test-cluster-name"

			mocks.serviceClient.
				EXPECT().
				List(ctx).
				Return(&corev1.ServiceList{
					Items: []corev1.Service{wrongService, rightService},
				}, nil)

			mocks.meshWorkloadClient.
				EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{Items: nil}, nil)

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, clients.ObjectMetaToObjectKey(mesh.ObjectMeta)).
				Return(mesh, nil)

			mocks.meshServiceClient.
				EXPECT().
				GetMeshService(ctx, client.ObjectKey{
					Name:      meshServiceName,
					Namespace: env.GetWriteNamespace(),
				}).
				Return(&v1alpha1.MeshService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      meshServiceName,
						Namespace: env.GetWriteNamespace(),
					},
					Spec: discovery_types.MeshServiceSpec{
						KubeService: &discovery_types.MeshServiceSpec_KubeService{
							Ref: &core_types.ResourceRef{
								Name:      rightService.GetName(),
								Namespace: rightService.GetNamespace(),
								Cluster:   clusterName,
							},
							WorkloadSelectorLabels: rightService.Spec.Selector,
							Ports: []*discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Name:     "port-1",
								Port:     80,
								Protocol: "TCP",
							}},
						},
						Mesh: meshWorkloadEvent.Spec.Mesh,
					},
				}, nil)

			err := mocks.meshWorkloadCallback(meshWorkloadEvent)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("service event", func() {
		It("can associate a service with an existing mesh workload", func() {
			mocks := setupMocks()

			mesh := &v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
				},
			}

			serviceEvent := &corev1.Service{
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app":   "test-app",
						"track": "canary",
					},
					Ports: []corev1.ServicePort{{
						Name:       "port-1",
						Protocol:   "TCP",
						Port:       80,
						TargetPort: intstr.IntOrString{IntVal: 8080},
						NodePort:   32000,
					}},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-svc",
					Namespace: "my-ns",
					Labels: map[string]string{
						"k1": "v1",
						"k2": "v2",
					},
				},
			}

			wrongWorkload := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":   "test-app",
							"track": "production",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}
			rightWorkloadV1 := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":     "test-app",
							"track":   "canary",
							"version": "v1",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}
			rightWorkloadV2 := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":     "test-app",
							"track":   "canary",
							"version": "v2",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      "istio-test-mesh",
						Namespace: "isito-system",
					},
				},
			}

			meshServiceName := "my-svc-my-ns-test-cluster-name"

			mocks.meshClient.
				EXPECT().
				GetMesh(ctx, client.ObjectKey{
					Name:      mesh.Name,
					Namespace: mesh.Namespace,
				}).
				Return(mesh, nil).
				Times(2)

			mocks.meshWorkloadClient.
				EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{
						*wrongWorkload,
						*rightWorkloadV1,
						*rightWorkloadV2,
					},
				}, nil)

			mocks.meshServiceClient.
				EXPECT().
				GetMeshService(ctx, client.ObjectKey{
					Name:      meshServiceName,
					Namespace: env.GetWriteNamespace(),
				}).
				Return(nil, errors.NewNotFound(v1alpha1.Resource("meshservice"), meshServiceName))

			mocks.meshServiceClient.
				EXPECT().
				CreateMeshService(ctx, &v1alpha1.MeshService{
					ObjectMeta: metav1.ObjectMeta{
						Name:      meshServiceName,
						Namespace: env.GetWriteNamespace(),
						Labels:    mesh_service.DiscoveryLabels(clusterName, serviceEvent.GetName(), serviceEvent.GetNamespace()),
					},
					Spec: discovery_types.MeshServiceSpec{
						KubeService: &discovery_types.MeshServiceSpec_KubeService{
							Ref: &core_types.ResourceRef{
								Name:      serviceEvent.GetName(),
								Namespace: serviceEvent.GetNamespace(),
								Cluster:   clusterName,
							},
							WorkloadSelectorLabels: serviceEvent.Spec.Selector,
							Labels:                 serviceEvent.GetLabels(),
							Ports: []*discovery_types.MeshServiceSpec_KubeService_KubeServicePort{{
								Name:     "port-1",
								Port:     80,
								Protocol: "TCP",
							}},
						},
						Mesh: rightWorkloadV1.Spec.Mesh,
						Subsets: map[string]*discovery_types.MeshServiceSpec_Subset{
							"version": {
								Values: []string{"v1", "v2"},
							},
						},
					},
				}).
				Return(nil)

			err := mocks.serviceCallback(serviceEvent)

			Expect(err).NotTo(HaveOccurred())
		})

		It("does not associate a service to any mesh workload if the labels don't match", func() {
			mocks := setupMocks()

			mesh := &v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: clusterName,
					},
				},
			}

			serviceEvent := &corev1.Service{
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app":   "test-app",
						"track": "canary",
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-svc",
					Namespace: "my-ns",
				},
			}

			wrongWorkload1 := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":   "test-app",
							"track": "production",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}
			wrongWorkload2 := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":   "test-other-unrelated-app",
							"track": "canary",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			mocks.meshWorkloadClient.
				EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{
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

			err := mocks.serviceCallback(serviceEvent)

			Expect(err).NotTo(HaveOccurred())
		})

		It("will bail out early if the mesh is on a different cluster than the service", func() {
			mocks := setupMocks()

			mesh := &v1alpha1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "istio-test-mesh",
					Namespace: "isito-system",
				},
				Spec: discovery_types.MeshSpec{
					Cluster: &core_types.ResourceRef{
						Name: "incorrect-cluster-name",
					},
				},
			}

			serviceEvent := &corev1.Service{
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app":   "test-app",
						"track": "canary",
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-svc",
					Namespace: "my-ns",
				},
			}

			wrongWorkload1 := &v1alpha1.MeshWorkload{
				Spec: discovery_types.MeshWorkloadSpec{
					KubeController: &discovery_types.MeshWorkloadSpec_KubeController{
						Labels: map[string]string{
							"app":   "test-app",
							"track": "production",
						},
					},
					Mesh: &core_types.ResourceRef{
						Name:      mesh.Name,
						Namespace: mesh.Namespace,
					},
				},
			}

			mocks.meshWorkloadClient.
				EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{
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

			err := mocks.serviceCallback(serviceEvent)

			Expect(err).NotTo(HaveOccurred())
		})

		It("bails out early if the mesh workload has no labels to match on", func() {
			mocks := setupMocks()

			serviceEvent := &corev1.Service{
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-svc",
					Namespace: "my-ns",
				},
			}

			err := mocks.serviceCallback(serviceEvent)

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
