package selector_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	kubernetes_apps "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1"
	networking_selector "github.com/solo-io/service-mesh-hub/pkg/selector"
	"github.com/solo-io/service-mesh-hub/services/common/constants"
	mock_mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager/mocks"
	mock_discovery "github.com/solo-io/service-mesh-hub/test/mocks/clients/discovery.zephyr.solo.io/v1alpha1"
	mock_kubernetes_apps "github.com/solo-io/service-mesh-hub/test/mocks/clients/kubernetes/apps/v1"
	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ResourceSelector", func() {
	var (
		ctrl                    *gomock.Controller
		ctx                     context.Context
		mockMeshServiceClient   *mock_discovery.MockMeshServiceClient
		mockMeshWorkloadClient  *mock_discovery.MockMeshWorkloadClient
		mockDynamicClientGetter *mock_mc_manager.MockDynamicClientGetter
		mockDeploymentClient    *mock_kubernetes_apps.MockDeploymentClient
		resourceSelector        networking_selector.ResourceSelector
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
		mockMeshServiceClient = mock_discovery.NewMockMeshServiceClient(ctrl)
		mockMeshWorkloadClient = mock_discovery.NewMockMeshWorkloadClient(ctrl)
		mockDynamicClientGetter = mock_mc_manager.NewMockDynamicClientGetter(ctrl)
		mockDeploymentClient = mock_kubernetes_apps.NewMockDeploymentClient(ctrl)
		resourceSelector = networking_selector.NewResourceSelector(
			mockMeshServiceClient,
			mockMeshWorkloadClient,
			func(client client.Client) kubernetes_apps.DeploymentClient {
				return mockDeploymentClient
			},
			mockDynamicClientGetter,
		)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetMeshServicesByServiceSelector", func() {
		var (
			namespace1   string
			namespace2   string
			cluster1     string
			cluster2     string
			meshService1 v1alpha1.MeshService
			meshService2 v1alpha1.MeshService
			meshService3 v1alpha1.MeshService
			meshService4 v1alpha1.MeshService
			meshService5 v1alpha1.MeshService
		)
		BeforeEach(func() {
			cluster1 = "cluster1"
			cluster2 = "cluster2"
			namespace1 = "namespace1"
			namespace2 = "namespace2"
			meshService1 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-1"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-1",
							Namespace: namespace1,
							Cluster:   cluster1,
						},
						Labels: map[string]string{"k1": "v1"},
					},
				}}
			meshService2 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-2"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-2",
							Namespace: namespace1,
							Cluster:   cluster2,
						},
						Labels: map[string]string{"k1": "v1"},
					},
				}}
			meshService3 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-3"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-3",
							Namespace: namespace2,
							Cluster:   cluster1,
						},
						Labels: map[string]string{"k1": "v1", "other": "label"},
					},
				}}
			meshService4 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-4"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-4",
							Namespace: "other-namespace",
							Cluster:   cluster2,
						},
						Labels: map[string]string{"k1": "v1"},
					},
				}}
			meshService5 = v1alpha1.MeshService{
				ObjectMeta: v1.ObjectMeta{Name: "mesh-service-5"},
				Spec: types.MeshServiceSpec{
					KubeService: &types.MeshServiceSpec_KubeService{
						Ref: &core_types.ResourceRef{
							Name:      "kube-service-5",
							Namespace: namespace1,
							Cluster:   cluster2,
						},
						Labels: map[string]string{"other": "label"},
					},
				}}
			mockMeshServiceClient.
				EXPECT().
				ListMeshService(ctx).
				Return(&v1alpha1.MeshServiceList{
					Items: []v1alpha1.MeshService{meshService1, meshService2, meshService3, meshService4, meshService5},
				}, nil)
		})

		It("should select Destinations by labels and namespaces", func() {
			selector := &core_types.ServiceSelector{
				ServiceSelectorType: &core_types.ServiceSelector_Matcher_{
					Matcher: &core_types.ServiceSelector_Matcher{
						Labels:     map[string]string{"k1": "v1"},
						Namespaces: []string{namespace1, namespace2},
						Clusters:   []string{cluster1},
					},
				},
			}
			expectedMeshServices := []*v1alpha1.MeshService{&meshService1, &meshService3}

			meshServices, err := resourceSelector.GetMeshServicesByServiceSelector(ctx, selector)
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
			selector := &core_types.ServiceSelector{
				ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
					ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
						Services: []*core_types.ResourceRef{
							{Name: objKey1.Name, Namespace: objKey1.Namespace, Cluster: cluster1},
							{Name: objKey2.Name, Namespace: objKey2.Namespace, Cluster: cluster1},
						},
					},
				},
			}
			expectedMeshServices := []*v1alpha1.MeshService{&meshService1, &meshService3}
			meshServices, err := resourceSelector.GetMeshServicesByServiceSelector(ctx, selector)
			Expect(err).ToNot(HaveOccurred())
			Expect(meshServices).To(ConsistOf(expectedMeshServices))
		})

		It("should return error if Service not found", func() {
			name := "non-existent-name"
			namespace := "non-existent-namespace"
			cluster := "non-existent-cluster"
			selector := &core_types.ServiceSelector{
				ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
					ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
						Services: []*core_types.ResourceRef{
							{Name: name, Namespace: namespace, Cluster: cluster},
						},
					},
				},
			}
			_, err := resourceSelector.GetMeshServicesByServiceSelector(ctx, selector)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.KubeServiceNotFound(name, namespace, cluster)))
		})

		It("should select across all namespaces and clusters", func() {
			selector := &core_types.ServiceSelector{
				ServiceSelectorType: &core_types.ServiceSelector_Matcher_{
					Matcher: &core_types.ServiceSelector_Matcher{
						Labels: map[string]string{"k1": "v1"},
					},
				},
			}
			expectedMeshServices := []*v1alpha1.MeshService{&meshService1, &meshService2, &meshService3, &meshService4}
			meshServices, err := resourceSelector.GetMeshServicesByServiceSelector(ctx, selector)
			Expect(err).ToNot(HaveOccurred())
			Expect(meshServices).To(ConsistOf(expectedMeshServices))
		})

		It("should select all services if selector ommitted", func() {
			selector := &core_types.ServiceSelector{}
			expectedMeshServices := []*v1alpha1.MeshService{&meshService1, &meshService2, &meshService3, &meshService4, &meshService5}
			meshServices, err := resourceSelector.GetMeshServicesByServiceSelector(ctx, selector)
			Expect(err).ToNot(HaveOccurred())
			Expect(meshServices).To(ConsistOf(expectedMeshServices))
		})
	})

	Describe("GetMeshWorkloadsByIdentitySelector", func() {
		It("selects everything when the given selector is nil", func() {
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
				},
			}
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)

			foundWorkloads, err := resourceSelector.GetMeshWorkloadsByIdentitySelector(ctx, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(foundWorkloads).To(HaveLen(2))
			Expect(foundWorkloads[0]).To(Equal(workload1))
			Expect(foundWorkloads[1]).To(Equal(workload2))
		})

		It("can select by matcher - namespace", func() {
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns1",
						},
					},
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns2",
						},
					},
				},
			}
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)

			foundWorkloads, err := resourceSelector.GetMeshWorkloadsByIdentitySelector(ctx, &core_types.IdentitySelector{
				IdentitySelectorType: &core_types.IdentitySelector_Matcher_{
					Matcher: &core_types.IdentitySelector_Matcher{
						Namespaces: []string{"ns2"},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(foundWorkloads).To(HaveLen(1))
			Expect(foundWorkloads[0]).To(Equal(workload2))
		})

		It("can select by matcher - cluster", func() {
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns1",
							Cluster:   "cluster-1",
						},
					},
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns2",
							Cluster:   "cluster-2",
						},
					},
				},
			}
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)

			foundWorkloads, err := resourceSelector.GetMeshWorkloadsByIdentitySelector(ctx, &core_types.IdentitySelector{
				IdentitySelectorType: &core_types.IdentitySelector_Matcher_{
					Matcher: &core_types.IdentitySelector_Matcher{
						Clusters: []string{"cluster-2"},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(foundWorkloads).To(HaveLen(1))
			Expect(foundWorkloads[0]).To(Equal(workload2))
		})

		It("can select by matcher - both namespace and cluster", func() {
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns1",
							Cluster:   "cluster-1",
						},
					},
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns2",
							Cluster:   "cluster-2",
						},
					},
				},
			}
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)

			foundWorkloads, err := resourceSelector.GetMeshWorkloadsByIdentitySelector(ctx, &core_types.IdentitySelector{
				IdentitySelectorType: &core_types.IdentitySelector_Matcher_{
					Matcher: &core_types.IdentitySelector_Matcher{
						Clusters:   []string{"cluster-2"},
						Namespaces: []string{"fake-namespace", "ns2"},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(foundWorkloads).To(HaveLen(1))
			Expect(foundWorkloads[0]).To(Equal(workload2))
		})

		It("can select by refs", func() {
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns1",
							Cluster:   "cluster-1",
						},
					},
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns2",
							Cluster:   "cluster-2",
						},
					},
				},
			}
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)

			foundWorkloads, err := resourceSelector.GetMeshWorkloadsByIdentitySelector(ctx, &core_types.IdentitySelector{
				IdentitySelectorType: &core_types.IdentitySelector_ServiceAccountRefs_{
					ServiceAccountRefs: &core_types.IdentitySelector_ServiceAccountRefs{
						ServiceAccounts: []*core_types.ResourceRef{{
							Namespace: "ns2",
							Cluster:   "cluster-2",
						}},
					},
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(foundWorkloads).To(HaveLen(1))
			Expect(foundWorkloads[0]).To(Equal(workload2))
		})
	})

	Describe("GetMeshWorkloadsByWorkloadSelector", func() {
		It("returns everything if the selector is nil", func() {
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns1",
							Cluster:   "cluster-1",
						},
					},
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns2",
							Cluster:   "cluster-2",
						},
					},
				},
			}
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)

			foundWorkloads, err := resourceSelector.GetMeshWorkloadsByIdentitySelector(ctx, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(foundWorkloads).To(HaveLen(2))
			Expect(foundWorkloads[0]).To(Equal(workload1))
			Expect(foundWorkloads[1]).To(Equal(workload2))
		})

		It("selects everything if neither labels nor namespaces is set", func() {
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns1",
							Cluster:   "cluster-1",
						},
					},
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns2",
							Cluster:   "cluster-2",
						},
					},
				},
			}
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)

			foundWorkloads, err := resourceSelector.GetMeshWorkloadsByWorkloadSelector(ctx, &core_types.WorkloadSelector{
				// intentionally empty
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(foundWorkloads).To(HaveLen(2))
			Expect(foundWorkloads[0]).To(Equal(workload1))
			Expect(foundWorkloads[1]).To(Equal(workload2))
		})

		It("can select by labels", func() {
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
					Labels: map[string]string{
						constants.CLUSTER: "cluster-1",
					},
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns1",
							Cluster:   "cluster-1",
							Name:      "controller-1",
						},
					},
				},
			}
			workload1Controller := &apps_v1.Deployment{
				ObjectMeta: v1.ObjectMeta{
					Name:      "controller-1",
					Namespace: "ns1",
					Labels: map[string]string{
						"deployment-label-1": "deployment-label-1-value",
					},
				},
			}
			cluster1 := &v1alpha1.KubernetesCluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "cluster-1",
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
					Labels: map[string]string{
						constants.CLUSTER: "cluster-2",
					},
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns2",
							Cluster:   "cluster-2",
							Name:      "controller-2",
						},
					},
				},
			}
			cluster2 := &v1alpha1.KubernetesCluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "cluster-2",
				},
			}
			workload2Controller := &apps_v1.Deployment{
				ObjectMeta: v1.ObjectMeta{
					Name:      "controller-2",
					Namespace: "ns2",
					Labels: map[string]string{
						"deployment-label-2": "deployment-label-2-value",
					},
				},
			}

			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)
			mockDynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, cluster1.GetName()).
				Return(nil, nil)
			mockDynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, cluster2.GetName()).
				Return(nil, nil)
			mockDeploymentClient.EXPECT().
				GetDeployment(ctx, client.ObjectKey{Name: workload1Controller.GetName(), Namespace: workload1Controller.GetNamespace()}).
				Return(workload1Controller, nil)
			mockDeploymentClient.EXPECT().
				GetDeployment(ctx, client.ObjectKey{Name: workload2Controller.GetName(), Namespace: workload2Controller.GetNamespace()}).
				Return(workload2Controller, nil)

			foundWorkloads, err := resourceSelector.GetMeshWorkloadsByWorkloadSelector(ctx, &core_types.WorkloadSelector{
				Labels: workload2Controller.Labels,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(foundWorkloads).To(HaveLen(1))
			Expect(foundWorkloads[0]).To(Equal(workload2))
		})

		It("can select by namespaces", func() {
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
					Labels: map[string]string{
						constants.CLUSTER: "cluster-1",
					},
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns1",
							Cluster:   "cluster-1",
							Name:      "controller-1",
						},
					},
				},
			}
			workload1Controller := &apps_v1.Deployment{
				ObjectMeta: v1.ObjectMeta{
					Name:      "controller-1",
					Namespace: "ns1",
					Labels: map[string]string{
						"deployment-label-1": "deployment-label-1-value",
					},
				},
			}
			cluster1 := &v1alpha1.KubernetesCluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "cluster-1",
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
					Labels: map[string]string{
						constants.CLUSTER: "cluster-2",
					},
				},
				Spec: types.MeshWorkloadSpec{
					KubeController: &types.MeshWorkloadSpec_KubeController{
						KubeControllerRef: &core_types.ResourceRef{
							Namespace: "ns2",
							Cluster:   "cluster-2",
							Name:      "controller-2",
						},
					},
				},
			}
			cluster2 := &v1alpha1.KubernetesCluster{
				ObjectMeta: v1.ObjectMeta{
					Name: "cluster-2",
				},
			}
			workload2Controller := &apps_v1.Deployment{
				ObjectMeta: v1.ObjectMeta{
					Name:      "controller-2",
					Namespace: "ns2",
					Labels: map[string]string{
						"deployment-label-2": "deployment-label-2-value",
					},
				},
			}

			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)
			mockDynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, cluster1.GetName()).
				Return(nil, nil)
			mockDynamicClientGetter.EXPECT().
				GetClientForCluster(ctx, cluster2.GetName()).
				Return(nil, nil)
			mockDeploymentClient.EXPECT().
				GetDeployment(ctx, client.ObjectKey{Name: workload1Controller.GetName(), Namespace: workload1Controller.GetNamespace()}).
				Return(workload1Controller, nil)
			mockDeploymentClient.EXPECT().
				GetDeployment(ctx, client.ObjectKey{Name: workload2Controller.GetName(), Namespace: workload2Controller.GetNamespace()}).
				Return(workload2Controller, nil)

			foundWorkloads, err := resourceSelector.GetMeshWorkloadsByWorkloadSelector(ctx, &core_types.WorkloadSelector{
				Namespaces: []string{"ns2"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(foundWorkloads).To(HaveLen(1))
			Expect(foundWorkloads[0]).To(Equal(workload2))
		})
	})

	Describe("GetMeshServiceByRefSelector", func() {
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
			mockMeshServiceClient.EXPECT().ListMeshService(ctx, destinationKey).Return(
				&v1alpha1.MeshServiceList{
					Items: []v1alpha1.MeshService{expectedMeshService}}, nil)
			meshService, err := resourceSelector.GetMeshServiceByRefSelector(ctx, serviceName, serviceNamespace, serviceCluster)
			Expect(err).ToNot(HaveOccurred())
			Expect(meshService).To(Equal(&expectedMeshService))
		})

		It("should return an error if multiple MeshServices found", func() {
			serviceName := "kubeServiceName"
			serviceNamespace := "kubeServiceNamespace"
			serviceCluster := "destinationClusterName"
			destinationKey := client.MatchingLabels(map[string]string{
				constants.KUBE_SERVICE_NAME:      serviceName,
				constants.KUBE_SERVICE_NAMESPACE: serviceNamespace,
				constants.CLUSTER:                serviceCluster,
			})
			mockMeshServiceClient.EXPECT().ListMeshService(ctx, destinationKey).Return(
				&v1alpha1.MeshServiceList{
					Items: []v1alpha1.MeshService{{}, {}}}, nil)
			_, err := resourceSelector.GetMeshServiceByRefSelector(ctx, serviceName, serviceNamespace, serviceCluster)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.MultipleMeshServicesFound(serviceName, serviceNamespace, serviceCluster)))
		})

		It("should return an error if multiple MeshServices found", func() {
			serviceName := "kubeServiceName"
			serviceNamespace := "kubeServiceNamespace"
			serviceCluster := "destinationClusterName"
			destinationKey := client.MatchingLabels(map[string]string{
				constants.KUBE_SERVICE_NAME:      serviceName,
				constants.KUBE_SERVICE_NAMESPACE: serviceNamespace,
				constants.CLUSTER:                serviceCluster,
			})
			mockMeshServiceClient.EXPECT().ListMeshService(ctx, destinationKey).Return(
				&v1alpha1.MeshServiceList{
					Items: []v1alpha1.MeshService{}}, nil)
			_, err := resourceSelector.GetMeshServiceByRefSelector(ctx, serviceName, serviceNamespace, serviceCluster)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.MeshServiceNotFound(serviceName, serviceNamespace, serviceCluster)))
		})
	})

	Describe("GetMeshWorkloadByRefSelector", func() {
		It("errors if the cluster name is not provided", func() {
			meshWorkload, err := resourceSelector.GetMeshWorkloadByRefSelector(ctx, "test-name", "test-namespace", "")
			Expect(meshWorkload).To(BeNil())
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.MustProvideClusterName(&core_types.ResourceRef{
				Name:      "test-name",
				Namespace: "test-namespace",
			})))
		})

		It("can find a meshworkload", func() {
			controllerName, controllerNamespace, cluster := "test-name", "test-namespace", "test-cluster"
			expectedWorkload := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload",
				},
			}
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.KUBE_CONTROLLER_NAME:      controllerName,
					constants.KUBE_CONTROLLER_NAMESPACE: controllerNamespace,
					constants.CLUSTER:                   cluster,
				}).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*expectedWorkload},
				}, nil)

			foundWorkload, err := resourceSelector.GetMeshWorkloadByRefSelector(ctx, controllerName, controllerNamespace, cluster)
			Expect(err).To(BeNil())
			Expect(foundWorkload).To(Equal(expectedWorkload))
		})

		It("returns an error if more than one mesh workload is found", func() {
			controllerName, controllerNamespace, cluster := "test-name", "test-namespace", "test-cluster"
			workload1 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-1",
				},
			}
			workload2 := &v1alpha1.MeshWorkload{
				ObjectMeta: v1.ObjectMeta{
					Name: "my-workload-2",
				},
			}
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.KUBE_CONTROLLER_NAME:      controllerName,
					constants.KUBE_CONTROLLER_NAMESPACE: controllerNamespace,
					constants.CLUSTER:                   cluster,
				}).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{*workload1, *workload2},
				}, nil)

			foundWorkload, err := resourceSelector.GetMeshWorkloadByRefSelector(ctx, controllerName, controllerNamespace, cluster)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.MultipleMeshWorkloadsFound(controllerName, controllerNamespace, cluster)))
			Expect(foundWorkload).To(BeNil())
		})

		It("returns the appropriate error if no mesh workload is found", func() {
			controllerName, controllerNamespace, cluster := "test-name", "test-namespace", "test-cluster"
			mockMeshWorkloadClient.EXPECT().
				ListMeshWorkload(ctx, client.MatchingLabels{
					constants.KUBE_CONTROLLER_NAME:      controllerName,
					constants.KUBE_CONTROLLER_NAMESPACE: controllerNamespace,
					constants.CLUSTER:                   cluster,
				}).
				Return(&v1alpha1.MeshWorkloadList{
					Items: []v1alpha1.MeshWorkload{},
				}, nil)

			foundWorkload, err := resourceSelector.GetMeshWorkloadByRefSelector(ctx, controllerName, controllerNamespace, cluster)
			Expect(err).To(testutils.HaveInErrorChain(networking_selector.MeshWorkloadNotFound(controllerName, controllerNamespace, cluster)))
			Expect(foundWorkload).To(BeNil())
		})
	})
})
