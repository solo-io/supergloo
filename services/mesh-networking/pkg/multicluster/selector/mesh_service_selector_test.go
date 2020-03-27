package selector_test

import (
	"context"

	types2 "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	mock_core "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery/mocks"
	"github.com/solo-io/mesh-projects/services/common/constants"
	networking_selector "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/multicluster/selector"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MeshServiceSelector", func() {
	var (
		ctrl                   *gomock.Controller
		ctx                    context.Context
		mockMeshServiceClient  *mock_core.MockMeshServiceClient
		meshServiceSelector    networking_selector.MeshServiceSelector
		managementPlaneCluster = "management-plane-cluster"
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshServiceClient = mock_core.NewMockMeshServiceClient(ctrl)
		meshServiceSelector = networking_selector.NewMeshServiceSelector(mockMeshServiceClient)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("MeshServiceSelector error states", func() {
		BeforeEach(func() {
			mockMeshServiceClient.
				EXPECT().
				List(ctx).
				Return(&v1alpha1.MeshServiceList{
					Items: []v1alpha1.MeshService{},
				}, nil)
		})

		It("return error if both Selector.labels/names/cluster and Refs specified", func() {
			selector := &core_types.Selector{
				Labels:     map[string]string{"k1": "v1"},
				Namespaces: []string{"namespace"},
				Refs: []*core_types.ResourceRef{
					{Name: "name", Namespace: "namespace"},
				},
			}
			_, err := meshServiceSelector.GetMatchingMeshServices(ctx, selector)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.InvalidSelectorErr))
		})

		It("return error if Selector.cluster specified", func() {
			selector := &core_types.Selector{
				Labels:     map[string]string{"k1": "v1"},
				Namespaces: []string{"namespace"},
				Cluster:    &types2.StringValue{Value: "remote-cluster"},
			}
			_, err := meshServiceSelector.GetMatchingMeshServices(ctx, selector)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.ClusterSelectorNotSupported))
		})
	})

	Describe("MeshService selection", func() {
		var (
			namespace1   string
			namespace2   string
			meshService1 v1alpha1.MeshService
			meshService2 v1alpha1.MeshService
			meshService3 v1alpha1.MeshService
			meshService4 v1alpha1.MeshService
			meshService5 v1alpha1.MeshService
		)
		BeforeEach(func() {
			namespace1 = "namespace1"
			namespace2 = "namespace2"
			meshService1 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-1"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-1",
							Namespace: namespace1,
							Cluster:   &types2.StringValue{Value: managementPlaneCluster},
						},
						Labels: map[string]string{"k1": "v1"},
					},
				}}
			meshService2 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-2"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-2",
							Namespace: namespace1,
						},
						Labels: map[string]string{"k1": "v1"},
					},
				}}
			meshService3 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-3"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-3",
							Namespace: namespace2,
							Cluster:   &types2.StringValue{Value: managementPlaneCluster},
						},
						Labels: map[string]string{"k1": "v1", "other": "label"},
					},
				}}
			meshService4 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-4"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-4",
							Namespace: "other-namespace",
						},
						Labels: map[string]string{"k1": "v1"},
					},
				}}
			meshService5 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-5"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-5",
							Namespace: namespace1,
						},
						Labels: map[string]string{"other": "label"},
					},
				}}
			mockMeshServiceClient.
				EXPECT().
				List(ctx).
				Return(&v1alpha1.MeshServiceList{
					Items: []v1alpha1.MeshService{meshService1, meshService2, meshService3, meshService4, meshService5},
				}, nil)
		})

		It("should select Destinations by labels and namespaces", func() {
			selector := &core_types.Selector{
				Labels:     map[string]string{"k1": "v1"},
				Namespaces: []string{namespace1, namespace2},
			}
			expectedMeshServices := []*v1alpha1.MeshService{&meshService1, &meshService2, &meshService3}

			meshServices, err := meshServiceSelector.GetMatchingMeshServices(ctx, selector)
			Expect(err).ToNot(HaveOccurred())
			Expect(meshServices).To(ConsistOf(expectedMeshServices))
		})

		It("should select by resource ref", func() {
			objKey1 := client.ObjectKey{
				Name:      meshService1.Spec.GetKubeService().GetRef().GetName(),
				Namespace: meshService1.Spec.GetKubeService().GetRef().GetNamespace(),
			}
			objKey2 := client.ObjectKey{
				Name:      meshService3.Spec.GetKubeService().GetRef().GetName(),
				Namespace: meshService3.Spec.GetKubeService().GetRef().GetNamespace(),
			}
			selector := &core_types.Selector{
				Refs: []*core_types.ResourceRef{
					{Name: objKey1.Name, Namespace: objKey1.Namespace, Cluster: &types2.StringValue{Value: managementPlaneCluster}},
					{Name: objKey2.Name, Namespace: objKey2.Namespace, Cluster: &types2.StringValue{Value: managementPlaneCluster}},
				},
			}
			expectedMeshServices := []*v1alpha1.MeshService{&meshService1, &meshService3}
			meshServices, err := meshServiceSelector.GetMatchingMeshServices(ctx, selector)
			Expect(err).ToNot(HaveOccurred())
			Expect(meshServices).To(ConsistOf(expectedMeshServices))
		})

		It("should return error if Service not found", func() {
			objKey1 := client.ObjectKey{
				Name:      "non-existent-name",
				Namespace: "non-existent-namespace",
			}
			selector := &core_types.Selector{
				Refs: []*core_types.ResourceRef{
					{Name: objKey1.Name, Namespace: objKey1.Namespace, Cluster: &types2.StringValue{Value: managementPlaneCluster}},
				},
			}
			_, err := meshServiceSelector.GetMatchingMeshServices(ctx, selector)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.KubeServiceNotFound(objKey1.Name, objKey1.Namespace)))
		})

		It("should select across all namespaces", func() {
			selector := &core_types.Selector{
				Labels: map[string]string{"k1": "v1"},
			}
			expectedMeshServices := []*v1alpha1.MeshService{&meshService1, &meshService2, &meshService3, &meshService4}
			meshServices, err := meshServiceSelector.GetMatchingMeshServices(ctx, selector)
			Expect(err).ToNot(HaveOccurred())
			Expect(meshServices).To(ConsistOf(expectedMeshServices))
		})
	})

	Describe("get backing MeshService", func() {
		It("should return MeshService if found", func() {
			serviceName := "kubeServiceName"
			serviceNamespace := "kubeServiceNamespace"
			serviceCluster := "destinationClusterName"
			destinationKey := client.MatchingLabels(map[string]string{
				constants.KUBE_SERVICE_NAME:      serviceName,
				constants.KUBE_SERVICE_NAMESPACE: serviceNamespace,
				constants.CLUSTER:                serviceCluster,
			})
			expectedMeshService := v1alpha1.MeshService{}
			mockMeshServiceClient.EXPECT().List(ctx, destinationKey).Return(
				&v1alpha1.MeshServiceList{
					Items: []v1alpha1.MeshService{expectedMeshService}}, nil)
			meshService, err := meshServiceSelector.GetBackingMeshService(ctx, serviceName, serviceNamespace, serviceCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(meshService).To(Equal(&expectedMeshService))
		})

		It("should throw error if multiple MeshServices found", func() {
			serviceName := "kubeServiceName"
			serviceNamespace := "kubeServiceNamespace"
			serviceCluster := "destinationClusterName"
			destinationKey := client.MatchingLabels(map[string]string{
				constants.KUBE_SERVICE_NAME:      serviceName,
				constants.KUBE_SERVICE_NAMESPACE: serviceNamespace,
				constants.CLUSTER:                serviceCluster,
			})
			mockMeshServiceClient.EXPECT().List(ctx, destinationKey).Return(
				&v1alpha1.MeshServiceList{
					Items: []v1alpha1.MeshService{{}, {}}}, nil)
			_, err := meshServiceSelector.GetBackingMeshService(ctx, serviceName, serviceNamespace, serviceCluster)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.MultipleMeshServicesFound(serviceName, serviceNamespace, serviceCluster)))
		})

		It("should throw error if multiple MeshServices found", func() {
			serviceName := "kubeServiceName"
			serviceNamespace := "kubeServiceNamespace"
			serviceCluster := "destinationClusterName"
			destinationKey := client.MatchingLabels(map[string]string{
				constants.KUBE_SERVICE_NAME:      serviceName,
				constants.KUBE_SERVICE_NAMESPACE: serviceNamespace,
				constants.CLUSTER:                serviceCluster,
			})
			mockMeshServiceClient.EXPECT().List(ctx, destinationKey).Return(
				&v1alpha1.MeshServiceList{
					Items: []v1alpha1.MeshService{}}, nil)
			_, err := meshServiceSelector.GetBackingMeshService(ctx, serviceName, serviceNamespace, serviceCluster)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.MeshServiceNotFound(serviceName, serviceNamespace, serviceCluster)))
		})
	})
})
