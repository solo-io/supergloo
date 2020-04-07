package exploration_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/explore/exploration"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	mock_zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/networking/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mock_selector "github.com/solo-io/service-mesh-hub/pkg/selector/mocks"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Resource explorer", func() {
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

	Describe("ExploreService", func() {
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
			exploredMeshService := &discovery_v1alpha1.MeshService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "expected-mesh-service",
					Namespace: env.DefaultWriteNamespace,
				},
			}

			accessControlPolices := &networking_v1alpha1.AccessControlPolicyList{
				Items: []networking_v1alpha1.AccessControlPolicy{
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
			trafficPolicies := &networking_v1alpha1.TrafficPolicyList{
				Items: []networking_v1alpha1.TrafficPolicy{
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
				Return(exploredMeshService, nil)
			accessControlPolicyClient.EXPECT().
				List(ctx).
				Return(accessControlPolices, nil)
			trafficPolicyClient.EXPECT().
				List(ctx).
				Return(trafficPolicies, nil)
			resourceSelector.EXPECT().
				GetMeshServicesByServiceSelector(ctx, wrongServiceSelector).
				Return([]*discovery_v1alpha1.MeshService{}, nil).
				Times(2)
			resourceSelector.EXPECT().
				GetMeshServicesByServiceSelector(ctx, correctServiceSelector).
				Return([]*discovery_v1alpha1.MeshService{exploredMeshService}, nil).
				Times(2)

			explorer := exploration.NewResourceExplorer(trafficPolicyClient, accessControlPolicyClient, resourceSelector)
			explorationResult, err := explorer.ExploreService(ctx, exploration.FullyQualifiedKubeResource{
				Name:        serviceName,
				Namespace:   serviceNs,
				ClusterName: serviceCluster,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(explorationResult).To(Equal(&exploration.ExplorationResult{
				Policies: &exploration.Policies{
					AccessControlPolicies: []*networking_v1alpha1.AccessControlPolicy{&accessControlPolices.Items[1]},
					TrafficPolicies:       []*networking_v1alpha1.TrafficPolicy{&trafficPolicies.Items[1]},
				},
			}))
		})
	})

	Describe("ExploreWorkload", func() {
		It("can find all the config that applies to a service", func() {
			trafficPolicyClient := mock_zephyr_networking.NewMockTrafficPolicyClient(ctrl)
			accessControlPolicyClient := mock_zephyr_networking.NewMockAccessControlPolicyClient(ctrl)
			resourceSelector := mock_selector.NewMockResourceSelector(ctrl)

			controllerName, controllerNs, controllerCluster := "controller-name", "controller-ns", "controller-cluster"
			exploredMeshWorkload := &discovery_v1alpha1.MeshWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mesh-workload",
					Namespace: env.DefaultWriteNamespace,
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
			accessControlPolices := &networking_v1alpha1.AccessControlPolicyList{
				Items: []networking_v1alpha1.AccessControlPolicy{
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
			trafficPolicies := &networking_v1alpha1.TrafficPolicyList{
				Items: []networking_v1alpha1.TrafficPolicy{
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
				Return(exploredMeshWorkload, nil)
			accessControlPolicyClient.EXPECT().
				List(ctx).
				Return(accessControlPolices, nil)
			trafficPolicyClient.EXPECT().
				List(ctx).
				Return(trafficPolicies, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByIdentitySelector(ctx, wrongIdentitySelector).
				Return([]*discovery_v1alpha1.MeshWorkload{}, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByIdentitySelector(ctx, correctIdentitySelector).
				Return([]*discovery_v1alpha1.MeshWorkload{exploredMeshWorkload}, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByWorkloadSelector(ctx, wrongWorkloadSelector).
				Return([]*discovery_v1alpha1.MeshWorkload{}, nil)
			resourceSelector.EXPECT().
				GetMeshWorkloadsByWorkloadSelector(ctx, correctWorkloadSelector).
				Return([]*discovery_v1alpha1.MeshWorkload{exploredMeshWorkload}, nil)

			explorer := exploration.NewResourceExplorer(trafficPolicyClient, accessControlPolicyClient, resourceSelector)
			explorationResult, err := explorer.ExploreWorkload(ctx, exploration.FullyQualifiedKubeResource{
				Name:        controllerName,
				Namespace:   controllerNs,
				ClusterName: controllerCluster,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(explorationResult).To(Equal(&exploration.ExplorationResult{
				Policies: &exploration.Policies{
					AccessControlPolicies: []*networking_v1alpha1.AccessControlPolicy{&accessControlPolices.Items[1]},
					TrafficPolicies:       []*networking_v1alpha1.TrafficPolicy{&trafficPolicies.Items[1]},
				},
			}))
		})
	})
})
