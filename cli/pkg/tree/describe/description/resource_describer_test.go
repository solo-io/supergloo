package description_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/describe/description"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/test/mocks/clients/networking.zephyr.solo.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Resource describer", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("DescribeService", func() {
		It("can find all the config that applies to a service", func() {
			trafficPolicyClient := mock_zephyr_networking.NewMockTrafficPolicyClient(ctrl)
			accessControlPolicyClient := mock_zephyr_networking.NewMockAccessControlPolicyClient(ctrl)
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)
			serviceName, serviceNs, serviceCluster := "test-svc", "test-ns", "test-cluster"
			correctServiceSelector := &core_types.ServiceSelector{
				ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
					ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
						Services: []*core_types.ResourceRef{{Name: serviceName, Namespace: serviceNs, Cluster: serviceCluster}},
					},
				},
			}
			wrongServiceSelector := &core_types.ServiceSelector{
				ServiceSelectorType: &core_types.ServiceSelector_ServiceRefs_{
					ServiceRefs: &core_types.ServiceSelector_ServiceRefs{
						Services: []*core_types.ResourceRef{{Name: "other-name", Namespace: "other-ns", Cluster: serviceCluster}},
					},
				},
			}
			describedMeshService := &zephyr_discovery.MeshService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "expected-mesh-service",
					Namespace: env.GetWriteNamespace(),
				},
			}

			accessControlPolices := &zephyr_networking.AccessControlPolicyList{
				Items: []zephyr_networking.AccessControlPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-1"},
						Spec: types.AccessControlPolicySpec{
							DestinationSelector: wrongServiceSelector,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-2"},
						Spec: types.AccessControlPolicySpec{
							DestinationSelector: correctServiceSelector,
						},
					},
				},
			}
			trafficPolicies := &zephyr_networking.TrafficPolicyList{
				Items: []zephyr_networking.TrafficPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "tp-1"},
						Spec: types.TrafficPolicySpec{
							DestinationSelector: wrongServiceSelector,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "tp-2"},
						Spec: types.TrafficPolicySpec{
							DestinationSelector: correctServiceSelector,
						},
					},
				},
			}

			resourceSelector.EXPECT().
				GetMeshServiceByRefSelector(ctx, serviceName, serviceNs, serviceCluster).
				Return(describedMeshService, nil)
			accessControlPolicyClient.EXPECT().
				ListAccessControlPolicy(ctx).
				Return(accessControlPolices, nil)
			trafficPolicyClient.EXPECT().
				ListTrafficPolicy(ctx).
				Return(trafficPolicies, nil)
			resourceSelector.EXPECT().
				GetMeshServicesByServiceSelector(ctx, wrongServiceSelector).
				Return([]*zephyr_discovery.MeshService{}, nil).
				Times(2)
			resourceSelector.EXPECT().
				GetMeshServicesByServiceSelector(ctx, correctServiceSelector).
				Return([]*zephyr_discovery.MeshService{describedMeshService}, nil).
				Times(2)

			describer := description.NewResourceDescriber(trafficPolicyClient, accessControlPolicyClient, resourceSelector)
			explorationResult, err := describer.DescribeService(ctx, description.FullyQualifiedKubeResource{
				Name:        serviceName,
				Namespace:   serviceNs,
				ClusterName: serviceCluster,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(explorationResult).To(Equal(&description.DescriptionResult{
				Policies: &description.Policies{
					AccessControlPolicies: []*zephyr_networking.AccessControlPolicy{&accessControlPolices.Items[1]},
					TrafficPolicies:       []*zephyr_networking.TrafficPolicy{&trafficPolicies.Items[1]},
				},
			}))
		})
	})

	Describe("DescribeWorkload", func() {
		It("can find all the config that applies to a service", func() {
			trafficPolicyClient := mock_zephyr_networking.NewMockTrafficPolicyClient(ctrl)
			accessControlPolicyClient := mock_zephyr_networking.NewMockAccessControlPolicyClient(ctrl)
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)

			controllerName, controllerNs, controllerCluster := "controller-name", "controller-ns", "controller-cluster"
			describedMeshWorkload := &zephyr_discovery.MeshWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mesh-workload",
					Namespace: env.GetWriteNamespace(),
				},
			}
			wrongIdentitySelector := &core_types.IdentitySelector{}
			correctIdentitySelector := &core_types.IdentitySelector{
				IdentitySelectorType: &core_types.IdentitySelector_ServiceAccountRefs_{},
			}
			wrongWorkloadSelector := &core_types.WorkloadSelector{}
			correctWorkloadSelector := &core_types.WorkloadSelector{
				Namespaces: []string{"doesn't-matter"},
			}
			accessControlPolices := &zephyr_networking.AccessControlPolicyList{
				Items: []zephyr_networking.AccessControlPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-1"},
						Spec: types.AccessControlPolicySpec{
							SourceSelector: wrongIdentitySelector,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "acp-2"},
						Spec: types.AccessControlPolicySpec{
							SourceSelector: correctIdentitySelector,
						},
					},
				},
			}
			trafficPolicies := &zephyr_networking.TrafficPolicyList{
				Items: []zephyr_networking.TrafficPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "tp-1"},
						Spec: types.TrafficPolicySpec{
							SourceSelector: wrongWorkloadSelector,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "tp-2"},
						Spec: types.TrafficPolicySpec{
							SourceSelector: correctWorkloadSelector,
						},
					},
				},
			}

			resourceSelector.EXPECT().
				GetMeshWorkloadByRefSelector(ctx, controllerName, controllerNs, controllerCluster).
				Return(describedMeshWorkload, nil)
			accessControlPolicyClient.EXPECT().
				ListAccessControlPolicy(ctx).
				Return(accessControlPolices, nil)
			trafficPolicyClient.EXPECT().
				ListTrafficPolicy(ctx).
				Return(trafficPolicies, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByIdentitySelector(ctx, wrongIdentitySelector).
				Return([]*zephyr_discovery.MeshWorkload{}, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByIdentitySelector(ctx, correctIdentitySelector).
				Return([]*zephyr_discovery.MeshWorkload{describedMeshWorkload}, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByWorkloadSelector(ctx, wrongWorkloadSelector).
				Return([]*zephyr_discovery.MeshWorkload{}, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByWorkloadSelector(ctx, correctWorkloadSelector).
				Return([]*zephyr_discovery.MeshWorkload{describedMeshWorkload}, nil)

			describer := description.NewResourceDescriber(trafficPolicyClient, accessControlPolicyClient, resourceSelector)
			explorationResult, err := describer.DescribeWorkload(ctx, description.FullyQualifiedKubeResource{
				Name:        controllerName,
				Namespace:   controllerNs,
				ClusterName: controllerCluster,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(explorationResult).To(Equal(&description.DescriptionResult{
				Policies: &description.Policies{
					AccessControlPolicies: []*zephyr_networking.AccessControlPolicy{&accessControlPolices.Items[1]},
					TrafficPolicies:       []*zephyr_networking.TrafficPolicy{&trafficPolicies.Items[1]},
				},
			}))
		})
	})
})
